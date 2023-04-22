package keeper

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	"github.com/golang/protobuf/proto" //nolint:staticcheck

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

var (
	// A valid IBC path for the LSM token must only consist of 1 channel hop along a transfer channel
	// (e.g. "transfer/channel-0")
	IsValidIBCPath = regexp.MustCompile(fmt.Sprintf(`^%s/(%s[0-9]{1,20})$`, transfertypes.PortID, channeltypes.ChannelPrefix)).MatchString

	// Timeout for the validator slash query that occurs at periodic deposit intervals
	SlashQueryTimeout = time.Minute * 5 // 5 minutes

	// Time for the IBC transfer of the LSM Token to the host zone
	LSMDepositTransferTimeout = time.Hour * 24 // 1 day

	// Time for the detokenization ICA
	DetokenizationTimeout = time.Hour * 24 // 1 day
)

// Parse the LSM Token's IBC denom hash into a DenomTrace object that contains the path and base denom
func (k Keeper) GetLSMTokenDenomTrace(ctx sdk.Context, denom string) (transfertypes.DenomTrace, error) {
	ibcPrefix := transfertypes.DenomPrefix + "/"

	// Confirm the LSM Token is a valid IBC token (has "ibc/" prefix)
	if !strings.HasPrefix(denom, ibcPrefix) {
		return transfertypes.DenomTrace{}, errorsmod.Wrapf(types.ErrInvalidLSMToken, "lsm token is not an IBC token (%s)", denom)
	}

	// Parse the hash string after the "ibc/" prefix into hex bytes
	hexHash := denom[len(ibcPrefix):]
	hash, err := transfertypes.ParseHexHash(hexHash)
	if err != nil {
		return transfertypes.DenomTrace{}, errorsmod.Wrapf(err, "unable to get ibc hex hash from denom %s", denom)
	}

	// Lookup the trace from the hash
	denomTrace, found := k.RecordsKeeper.TransferKeeper.GetDenomTrace(ctx, hash)
	if !found {
		return transfertypes.DenomTrace{}, errorsmod.Wrapf(types.ErrInvalidLSMToken, "denom trace not found for %s", denom)
	}

	return denomTrace, nil
}

// Parses the LSM token's IBC path (e.g. transfer/channel-0) and confirms the channel ID matches
// the transfer channel of a supported host zone
func (k Keeper) GetHostZoneFromLSMTokenPath(ctx sdk.Context, path string) (types.HostZone, error) {
	// Validate path regex which confirms the token originated only one hop away (e.g. transfer/channel-0)
	if !IsValidIBCPath(path) {
		return types.HostZone{}, errorsmod.Wrapf(types.ErrInvalidLSMToken, "ibc path of LSM token (%s) cannot be more than 1 hop away", path)
	}

	// Remove the "transfer/" prefix
	channelId := strings.ReplaceAll(path, transfertypes.PortID+"/", "")

	// Confirm the channel is from one of Stride's supported host zones
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if hostZone.TransferChannelId == channelId {
			return hostZone, nil
		}
	}

	return types.HostZone{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
		"transfer channel-id from LSM token (%s) does not match any registered host zone", channelId)
}

// Parses the LSM token's denom (of the form {validatorAddress}/{recordId}) and confirms that the validator
// is in the Stride validator set
func (k Keeper) GetValidatorFromLSMTokenDenom(denom string, validators []*types.Validator) (types.Validator, error) {
	// Denom is of the form {validatorAddress}/{recordId}
	split := strings.Split(denom, "/")
	if len(split) != 2 {
		return types.Validator{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
			"lsm token base denom is not of the format {val-address}/{record-id} (%s)", denom)
	}
	validatorAddress := split[0]

	// Confirm validator is in Stride's validator set
	for _, validator := range validators {
		if validator.Address == validatorAddress {
			return *validator, nil
		}
	}

	return types.Validator{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
		"validator (%s) is not registered in the Stride validator set", validatorAddress)
}

// Checks if we need to issue an ICQ to check if a validator was slashed
// The query runs at periodic intervals defined by the ValidatorSlashQueryInterval
func (k Keeper) ShouldCheckIfValidatorWasSlashed(ctx sdk.Context, validator types.Validator, stakeAmount sdkmath.Int) bool {
	params := k.GetParams(ctx)
	queryInterval := sdk.NewIntFromUint64(params.ValidatorSlashQueryInterval)

	// If the query interval is disabled, do not submit a query
	// This should not be possible with the current parameter validation
	//  function which enforces that it's greater than 0
	if queryInterval.IsZero() {
		return false
	}

	oldProgress := validator.SlashQueryProgressTracker
	newProgress := validator.SlashQueryProgressTracker.Add(stakeAmount)

	// Submit query if the query interval checkpoint has been breached
	// Ex: Query Interval: 1000, Old Progress: 900, New Progress: 1100
	//     => OldProgress/Interval: 0, NewProgress/Interval: 1
	return oldProgress.Quo(queryInterval).LT(newProgress.Quo(queryInterval))
}

// Helper function to Refund an LSM Token to the user if both:
//  a) the LSMLiquidStake transaction is interrupted with a validator exchange rate query, and
//  b) the handling of the query callback fails
func (k Keeper) RefundLSMToken(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake) error {
	liquidStakerAddress := lsmLiquidStake.Staker
	hostZoneAddress, err := sdk.AccAddressFromBech32(lsmLiquidStake.HostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "host zone address is invalid")
	}
	refundToken := lsmLiquidStake.LSMIBCToken

	if err := k.bankKeeper.SendCoins(ctx, hostZoneAddress, liquidStakerAddress, sdk.NewCoins(refundToken)); err != nil {
		return errorsmod.Wrapf(err,
			"failed to send tokens from Module Account (%s) to Staker (%s)", hostZoneAddress.String(), liquidStakerAddress.String())
	}

	return nil
}

// Submits an ICA to "Redeem" an LSM Token - meaning converting the token into native stake
// This function is called in the EndBlocker which means if the ICA submission fails,
//   any modified state is not reverted
// The deposit Status is intentionally updated before the ICA is submitted even though it will NOT be reverted
//   if the ICA fails to send. This is because a failure is likely caused by a closed ICA channel, and the status
//   update will prevent the ICA from being continuously re-submitted. When the ICA channel is restored, the
//   deposit status will get reset, and the ICA will be attempted again.
func (k Keeper) DetokenizeLSMDeposit(ctx sdk.Context, hostZone types.HostZone, deposit types.LSMTokenDeposit) error {
	// Get the delegation account (which owns the LSM token)
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Build the detokenization ICA message
	token := sdk.NewCoin(deposit.Denom, deposit.Amount)
	detokenizeMsg := []sdk.Msg{&types.MsgRedeemTokensforShares{
		DelegatorAddress: hostZone.DelegationIcaAddress,
		Amount:           token,
	}}

	// Store the LSMTokenDeposit for the callback
	callbackArgs := types.DetokenizeSharesCallback{
		Deposit: &deposit,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	if err != nil {
		return err
	}

	// Mark the deposit as IN_PROGRESS
	deposit.Status = types.DETOKENIZATION_IN_PROGRESS
	k.SetLSMTokenDeposit(ctx, deposit)

	// Submit the ICA with a 24 hour timeout
	timeout := uint64(ctx.BlockTime().UnixNano() + (DetokenizationTimeout).Nanoseconds()) // 1 day
	if _, err := k.SubmitTxs(ctx, hostZone.ConnectionId, detokenizeMsg, types.ICAAccountType_DELEGATION, timeout, ICACallbackID_Detokenize, callbackArgsBz); err != nil {
		return errorsmod.Wrapf(err, "unable to submit detokenization ICA for %s", deposit.Denom)
	}

	return nil
}

// Loops through all active host zones, grabs the queued LSMTokenDeposits for that host
// that are in status DETOKENIZATION_QUEUE, and submits the detokenization ICA for each
func (k Keeper) DetokenizeAllLSMDeposits(ctx sdk.Context) {
	// Submit detokenization ICAs for each active host zone
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		// Get the host zone's delegation ICA portID
		delegationICAOwner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
		delegationICAPortID, err := icatypes.NewControllerPortID(delegationICAOwner)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to get delegation port ID for %s: %s", hostZone.ChainId, err))
		}

		// If the delegation channel is not open, skip this host zone
		_, isOpen := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, hostZone.ConnectionId, delegationICAPortID)
		if !isOpen {
			k.Logger(ctx).Error(fmt.Sprintf("Skipping detokenization ICAs for %s - Delegation ICA channel is closed", hostZone.ChainId))
			continue
		}

		// If the delegation channel is open, submit the detokenize ICA
		for _, deposit := range k.GetLSMDepositsForHostZoneWithStatus(ctx, hostZone.ChainId, types.DETOKENIZATION_QUEUE) {
			if err := k.DetokenizeLSMDeposit(ctx, hostZone, deposit); err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Unable to submit detokenization ICAs for %v%s on %s: %s",
					deposit.Amount, deposit.Denom, hostZone.ChainId, err.Error()))
				continue
			}
			k.Logger(ctx).Info(fmt.Sprintf("Submitted detokenization ICA for deposit %v%s on %s", deposit.Amount, deposit.Denom, hostZone.ChainId))
		}
	}
}

// When receiving LSM Liquid Stakes, the distribution of stake from these tokens
//   will be inconsistent with the host zone's validator set
// This portion of stake is designated as the "Unbalanced Delegation"
// This function rebalances the unbalanced portion according to the validator weights,
//   and should not be confused with the main Rebalance function which rebalances
//   the "Balanced Delegation"
// Note: this cannot be run more than once in a single unbonding period
func (k Keeper) RebalanceTokenizedDeposits(ctx sdk.Context, dayNumber uint64) {
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		numRebalance := uint64(len(hostZone.Validators))

		// TODO [LSM]: Uncomment once UnbondingPeriod is added to the host zone
		// if dayNumber%hostZone.UnbondingPeriod != 0 {
		// 	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		// 		"Host does not rebalance this epoch (Unbonding Period: %d, Epoch: %d)", hostZone.UnbondingPeriod, dayNumber))
		// 	continue
		// }

		if err := k.RebalanceDelegations(ctx, hostZone.ChainId, types.DelegationType_UNBALANCED, numRebalance); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to rebalance delegations for %s: %s", hostZone.ChainId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(fmt.Sprintf("Successfully rebalanced UnbalancedDelegations for %s", hostZone.ChainId))
	}
}
