package keeper

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

var (
	// A valid IBC path for the LSM token must only consist of 1 channel hop along a transfer channel
	// (e.g. "transfer/channel-0")
	IsValidIBCPath = regexp.MustCompile(fmt.Sprintf(`^%s/(%s[0-9]{1,20})$`, transfertypes.PortID, channeltypes.ChannelPrefix)).MatchString

	// Timeout for the validator slash query that occurs at periodic deposit intervals
	LSMSlashQueryTimeout = time.Minute * 5 // 5 minutes

	// Time for the detokenization ICA
	DetokenizationTimeout = time.Hour * 24 // 1 day
)

// Validates the parameters supplied with this LSMLiquidStake, including that the denom
// corresponds with a valid LSM Token and that the user has sufficient balance
//
// This is called once at the beginning of the liquid stake, and is, potentially, called
// again at the end (if the transaction was asynchronous due to an intermediate slash query)
//
// This function returns the associated host zone and validator along with the initial deposit record
func (k Keeper) ValidateLSMLiquidStake(ctx sdk.Context, msg types.MsgLSMLiquidStake) (types.LSMLiquidStake, error) {
	// Get the denom trace from the IBC hash - this includes the full path and base denom
	// Ex: LSMTokenIbcDenom of `ibc/XXX` might create a DenomTrace with:
	//     BaseDenom: cosmosvaloperXXX/42, Path: transfer/channel-0
	denomTrace, err := k.GetLSMTokenDenomTrace(ctx, msg.LsmTokenIbcDenom)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}

	// Get the host zone and validator address from the path and base denom respectively
	lsmTokenBaseDenom := denomTrace.BaseDenom
	hostZone, err := k.GetHostZoneFromLSMTokenPath(ctx, denomTrace.Path)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}
	validator, err := k.GetValidatorFromLSMTokenDenom(lsmTokenBaseDenom, hostZone.Validators)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}

	// Confirm the staker has a sufficient balance to execute the liquid stake
	liquidStakerAddress := sdk.MustAccAddressFromBech32(msg.Creator)
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, msg.LsmTokenIbcDenom).Amount
	if balance.LT(msg.Amount) {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance)
	}

	// Build the LSMTokenDeposit record
	// The stToken will be added outside of this function
	depositId := GetLSMTokenDepositId(ctx.BlockHeight(), hostZone.ChainId, msg.Creator, lsmTokenBaseDenom)
	lsmTokenDeposit := recordstypes.LSMTokenDeposit{
		DepositId:        depositId,
		ChainId:          hostZone.ChainId,
		Denom:            lsmTokenBaseDenom,
		IbcDenom:         msg.LsmTokenIbcDenom,
		StakerAddress:    msg.Creator,
		ValidatorAddress: validator.Address,
		Amount:           msg.Amount,
		Status:           recordstypes.LSMTokenDeposit_DEPOSIT_PENDING,
	}

	// Return the wrapped deposit object with additional context (host zone and validator)
	return types.LSMLiquidStake{
		Deposit:   &lsmTokenDeposit,
		HostZone:  &hostZone,
		Validator: &validator,
	}, nil
}

// Generates a unique ID for the LSM token deposit so that, if a slash query is issued,
// the query callback can be joined back with this tx
// The key in the store for an LSMTokenDeposit is chainId + denom (meaning, there
// can only be 1 LSMLiquidStake in progress per tokenization)
func GetLSMTokenDepositId(blockHeight int64, chainId, stakerAddress, denom string) string {
	id := fmt.Sprintf("%d-%s-%s-%s", blockHeight, chainId, stakerAddress, denom)
	return fmt.Sprintf("%x", crypto.Sha256([]byte(id)))
}

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
		return types.HostZone{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
			"ibc path of LSM token (%s) cannot be more than 1 hop away", path)
	}

	// Remove the "transfer/" prefix
	channelId := strings.ReplaceAll(path, transfertypes.PortID+"/", "")

	// Confirm the channel is from one of Stride's supported host zones
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if hostZone.TransferChannelId == channelId {
			if !hostZone.LsmLiquidStakeEnabled {
				return hostZone, types.ErrLSMLiquidStakeDisabledForHostZone.Wrapf(
					"LSM liquid stake disabled for %s", hostZone.ChainId)
			}
			return hostZone, nil
		}
	}

	return types.HostZone{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
		"transfer channel-id from LSM token (%s) does not match any registered host zone", channelId)
}

// Parses the LSM token's denom (of the form {validatorAddress}/{recordId}) and confirms that the validator
// is in the Stride validator set and does not have an active slash query
func (k Keeper) GetValidatorFromLSMTokenDenom(denom string, validators []*types.Validator) (types.Validator, error) {
	// Denom is of the form {validatorAddress}/{recordId}
	split := strings.Split(denom, "/")
	if len(split) != 2 {
		return types.Validator{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
			"lsm token base denom is not of the format {val-address}/{record-id} (%s)", denom)
	}
	validatorAddress := split[0]

	// Confirm the validator:
	//  1. Is registered on Stride
	//  2. Does not have an active slash query in flight
	//  3. Has a known sharesToTokens rate
	for _, validator := range validators {
		if validator.Address == validatorAddress {
			if validator.SlashQueryInProgress {
				return types.Validator{}, errorsmod.Wrapf(types.ErrValidatorWasSlashed,
					"validator %s was slashed, liquid stakes from this validator are temporarily unavailable", validator.Address)
			}
			if validator.SharesToTokensRate.IsNil() || validator.SharesToTokensRate.IsZero() {
				return types.Validator{}, errorsmod.Wrapf(types.ErrValidatorSharesToTokensRateNotKnown,
					"validator %s sharesToTokens rate is not known", validator.Address)
			}
			return *validator, nil
		}
	}

	return types.Validator{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
		"validator (%s) is not registered in the Stride validator set", validatorAddress)
}

// Given an LSMToken representing a number of delegator shares, returns the stToken coin
// using the validator's sharesToTokens rate and the host zone redemption rate
//
//	StTokens = LSMTokenShares * Validator SharesToTokens Rate / Redemption Rate
//
// Note: in the event of a slash query, these tokens will be minted only if the
// validator's sharesToTokens rate did not change
func (k Keeper) CalculateLSMStToken(liquidStakedShares sdkmath.Int, lsmLiquidStake types.LSMLiquidStake) sdk.Coin {
	hostZone := lsmLiquidStake.HostZone
	validator := lsmLiquidStake.Validator

	liquidStakedTokens := sdk.NewDecFromInt(liquidStakedShares).Mul(validator.SharesToTokensRate)
	stAmount := (liquidStakedTokens.Quo(hostZone.RedemptionRate)).TruncateInt()

	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stCoin := sdk.NewCoin(stDenom, stAmount)

	return stCoin
}

// Determines the new slash query checkpoint, by mulitplying the query threshold percent by the current TVL
func (k Keeper) GetUpdatedSlashQueryCheckpoint(ctx sdk.Context, totalDelegations sdkmath.Int) sdkmath.Int {
	params := k.GetParams(ctx)
	queryThreshold := sdk.NewDecWithPrec(int64(params.ValidatorSlashQueryThreshold), 2) // percentage
	checkpoint := queryThreshold.Mul(sdk.NewDecFromInt(totalDelegations)).TruncateInt()
	return checkpoint
}

// Checks if we need to issue an ICQ to check if a validator was slashed
// The query runs at periodic intervals defined by the ValidatorSlashQueryInterval
// The interval is represented as percent of TVL
// (e.g. 1% means every LS that causes the progress to breach 1% of TVL triggers the query)
func (k Keeper) ShouldCheckIfValidatorWasSlashed(
	ctx sdk.Context,
	validator types.Validator,
	transactionStakeAmount sdkmath.Int,
) bool {
	// If the checkpoint is zero - that means either the threshold parameter is 0
	// (which should not be possible), or that the total host zone stake is 0
	// In either case, do not submit the query
	if validator.SlashQueryCheckpoint.IsZero() {
		return false
	}

	oldInterval := validator.SlashQueryProgressTracker.Quo(validator.SlashQueryCheckpoint)
	newInterval := validator.SlashQueryProgressTracker.Add(transactionStakeAmount).Quo(validator.SlashQueryCheckpoint)

	// Submit query if the query interval checkpoint has been breached
	// Ex: Query Threshold: 1%, TVL: 100k => 1k Checkpoint
	//     Old Progress Tracker: 900, Old Interval: 900 / 1000 => Interval 0,
	//     Stake: 200, New Progress Tracker: 1100, New Interval: 1100 / 1000 = 1.1 = 1
	//     => OldInterval: 0, NewInterval: 1 => Issue Slash Query
	return oldInterval.LT(newInterval)
}

// Loops through all active host zones, grabs queued LSMTokenDeposits for that host
// that are in status TRANSFER_QUEUE, and submits the IBC Transfer to the host
func (k Keeper) TransferAllLSMDeposits(ctx sdk.Context) {
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		// Ignore hosts that have not been successfully registered
		if hostZone.DelegationIcaAddress == "" {
			continue
		}

		// Submit an IBC transfer for all queued deposits
		queuedDeposits := k.RecordsKeeper.GetLSMDepositsForHostZoneWithStatus(
			ctx,
			hostZone.ChainId,
			recordstypes.LSMTokenDeposit_TRANSFER_QUEUE,
		)
		for _, deposit := range queuedDeposits {

			// If the IBC transfer fails to get off the ground, flag the deposit as FAILED
			// This is highly unlikely and would indicate a larger problem
			if err := k.RecordsKeeper.IBCTransferLSMToken(
				ctx,
				deposit,
				hostZone.TransferChannelId,
				hostZone.DepositAddress,
				hostZone.DelegationIcaAddress,
			); err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Unable to submit IBC Transfer of LSMToken for %v%s on %s: %s",
					deposit.Amount, deposit.Denom, hostZone.ChainId, err.Error()))
				k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, deposit, recordstypes.LSMTokenDeposit_TRANSFER_FAILED)
				continue
			}
			k.Logger(ctx).Info(fmt.Sprintf("Submitted IBC Transfer for LSM deposit %v%s on %s",
				deposit.Amount, deposit.Denom, hostZone.ChainId))

			k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, deposit, recordstypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS)
		}
	}
}

// Submits an ICA to "Redeem" an LSM Token - meaning converting the token into native stake
// This function is called in the EndBlocker which means if the ICA submission fails,
// any modified state is not reverted
//
// The deposit Status is intentionally updated before the ICA is submitted even though it will NOT be reverted
// if the ICA fails to send. This is because a failure is likely caused by a closed ICA channel, and the status
// update will prevent the ICA from being continuously re-submitted. When the ICA channel is restored, the
// deposit status will get reset, and the ICA will be attempted again.
func (k Keeper) DetokenizeLSMDeposit(ctx sdk.Context, hostZone types.HostZone, deposit recordstypes.LSMTokenDeposit) error {
	// Get the delegation account (which owns the LSM token)
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Build the detokenization ICA message
	token := sdk.NewCoin(deposit.Denom, deposit.Amount)
	detokenizeMsg := []proto.Message{&types.MsgRedeemTokensForShares{
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

	// Submit the ICA with a coonservative timeout
	timeout := uint64(ctx.BlockTime().UnixNano() + (DetokenizationTimeout).Nanoseconds())
	if _, err := k.SubmitTxs(
		ctx,
		hostZone.ConnectionId,
		detokenizeMsg,
		types.ICAAccountType_DELEGATION,
		timeout,
		ICACallbackID_Detokenize,
		callbackArgsBz,
	); err != nil {
		return errorsmod.Wrapf(err, "unable to submit detokenization ICA for %s", deposit.Denom)
	}

	// Mark the deposit as IN_PROGRESS
	k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, deposit, recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS)

	// Update the validator to say that it has a delegation change in progress
	if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, deposit.ValidatorAddress); err != nil {
		return err
	}
	k.SetHostZone(ctx, hostZone)

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
			continue
		}

		// If the delegation channel is not open, skip this host zone
		_, isOpen := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, hostZone.ConnectionId, delegationICAPortID)
		if !isOpen {
			k.Logger(ctx).Error(fmt.Sprintf("Skipping detokenization ICAs for %s - Delegation ICA channel is closed", hostZone.ChainId))
			continue
		}

		// If the delegation channel is open, submit the detokenize ICA
		queuedDeposits := k.RecordsKeeper.GetLSMDepositsForHostZoneWithStatus(
			ctx,
			hostZone.ChainId,
			recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		)
		for _, deposit := range queuedDeposits {
			if err := k.DetokenizeLSMDeposit(ctx, hostZone, deposit); err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Unable to submit detokenization ICAs for %v%s on %s: %s",
					deposit.Amount, deposit.Denom, hostZone.ChainId, err.Error()))
				continue
			}
			k.Logger(ctx).Info(fmt.Sprintf("Submitted detokenization ICA for deposit %v%s on %s", deposit.Amount, deposit.Denom, hostZone.ChainId))
		}
	}
}
