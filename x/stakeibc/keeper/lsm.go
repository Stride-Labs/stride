package keeper

import (
	"fmt"
	"regexp"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

var (
	// QUESTION: Is this too restrictive? Would we ever want to accept a LSM token from more than 1 hop away?
	// A valid IBC path for the LSM token must only consist of 1 channel hop along a transfer channel
	// (e.g. "transfer/channel-0")
	IsValidIBCPath = regexp.MustCompile(fmt.Sprintf(`^%s/(%s[0-9]{1,20})$`, transfertypes.PortID, channeltypes.ChannelPrefix)).MatchString
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
// The query runs at periodic intervals defined by the ValidatorExchangeRateQueryInterval
func (k Keeper) ShouldQueryValidatorExchangeRate(ctx sdk.Context, validator types.Validator, stakeAmount sdk.Int) bool {
	params := k.GetParams(ctx)
	queryInterval := sdk.NewIntFromUint64(params.ValidatorExchangeRateQueryInterval)

	// If the query interval is disabled, do not submit a query
	// This should not be possible with the current parameter validation
	//  function which enforces that it's greater than 0
	if queryInterval.IsZero() {
		return false
	}

	oldProgress := validator.ProgressTowardsExchangeRateQuery
	newProgress := validator.ProgressTowardsExchangeRateQuery.Add(stakeAmount)

	// Submit query if the query interval checkpoint has been breached
	return oldProgress.Quo(queryInterval).LT(newProgress.Quo(queryInterval))
}

// Helper function to Refund an LSM Token to the user if both:
//  a) the LSMLiquidStake transaction is interrupted with a validator exchange rate query, and
//  b) the handling of the query callback fails
func (k Keeper) RefundLSMToken(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake) error {
	liquidStakerAddress := lsmLiquidStake.Staker
	hostZoneAddress, err := sdk.AccAddressFromBech32(lsmLiquidStake.HostZone.Address)
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
