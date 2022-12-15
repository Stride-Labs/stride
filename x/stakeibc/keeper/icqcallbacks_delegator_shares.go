package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// DelegatorSharesCallback is a callback handler for UpdateValidatorSharesExchRate queries.
//
// In an attempt to get the ICA's delegation amount on a given validator, we have to query:
//  1. the validator's internal exchange rate
//  2. the Delegation ICA's delegated shares
//     And apply the following equation:
//     num_tokens = exchange_rate * num_shares
//
// This is the callback from query #2
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func DelegatorSharesCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	hostZone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		errMsg := fmt.Sprintf("no registered zone for queried chain ID (%s)", query.GetChainId())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Unmarshal the query response which returns a delegation object for the delegator/validator pair
	queriedDelgation := stakingtypes.Delegation{}
	err := k.cdc.Unmarshal(args, &queriedDelgation)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal queried delegation info for zone %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: HostZone: %s, Delegator: %s, Validator: %s, Shares: %v",
		hostZone.ChainId, queriedDelgation.DelegatorAddress, queriedDelgation.ValidatorAddress, queriedDelgation.Shares))

	// ensure ICQ can be issued now! else fail the callback
	isWithinWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrOutsideIcqWindow, errMsg)
	} else if !isWithinWindow {
		k.Logger(ctx).Error("delegator shares callback is outside ICQ window")
		return nil
	}

	// Grab the validator object form the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelgation.ValidatorAddress)
	if !found {
		errMsg := fmt.Sprintf("no registered validator for address (%s)", queriedDelgation.ValidatorAddress)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrValidatorNotFound, errMsg)
	}

	// get the validator's internal exchange rate, aborting if it hasn't been updated this epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	if validator.InternalExchangeRate.EpochNumber != strideEpochTracker.GetEpochNumber() {
		errMsg := fmt.Sprintf("DelegationCallback: validator (%s) internal exchange rate has not been updated this epoch (epoch #%d)",
			validator.Address, strideEpochTracker.GetEpochNumber())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	// TODO: make sure conversion math precision matches the sdk's staking module's version (we did it slightly differently)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	validatorTokens := queriedDelgation.Shares.Mul(validator.InternalExchangeRate.InternalTokensToSharesRate).TruncateInt()
	k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: HostZone: %s, Validator: %s, Previous NumTokens: %v, Current NumTokens: %v",
		hostZone.ChainId, validator.Address, validator.DelegationAmt, validatorTokens))

	// Confirm the validator has actually been slashed
	if validatorTokens.Equal(validator.DelegationAmt) {
		k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: Validator (%s) was not slashed", validator.Address))
		return nil
	} else if validatorTokens.GT(validator.DelegationAmt) {
		errMsg := fmt.Sprintf("DelegationCallback: Validator (%s) tokens returned from query is greater than the DelegationAmt", validator.Address)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	// TODO(TESTS-171) add some safety checks here (e.g. we could query the slashing module to confirm the decr in tokens was due to slash)
	// update our records of the total stakedbal and of the validator's delegation amt
	// NOTE:  we assume any decrease in delegation amt that's not tracked via records is a slash

	// Get slash percentage
	slashAmount := validator.DelegationAmt.Sub(validatorTokens)

	weight, err := cast.ToInt64E(validator.Weight)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert validator weight to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}

	slashPct := sdk.NewDecFromInt(slashAmount).Quo(sdk.NewDecFromInt(validator.DelegationAmt))
	k.Logger(ctx).Info(fmt.Sprintf("ICQ'd Delegation Amount Mismatch, HostZone: %s, Validator: %s, Delegator: %s, Records Tokens: %v, Tokens from ICQ %v, Slash Amount: %v, Slash Pct: %v!",
		hostZone.ChainId, validator.Address, queriedDelgation.DelegatorAddress, validator.DelegationAmt, validatorTokens, slashAmount, slashPct))

	// Abort if the slash was greater than 10%
	tenPercent := sdk.NewDec(10).Quo(sdk.NewDec(100))
	if slashPct.GT(tenPercent) {
		errMsg := fmt.Sprintf("DelegationCallback: Validator (%s) slashed but ABORTING update, slash is greater than 0.10 (%d)", validator.Address, slashPct)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrSlashGtTenPct, errMsg)
	}

	// Update the host zone and validator to reflect the weight and delegation change
	weightAdjustment := sdk.NewDecFromInt(validatorTokens).Quo(sdk.NewDecFromInt(validator.DelegationAmt))
	validator.Weight = sdk.NewDec(int64(weight)).Mul(weightAdjustment).TruncateInt().Uint64()
	validator.DelegationAmt = validator.DelegationAmt.Sub(slashAmount)

	hostZone.StakedBal = hostZone.StakedBal.Sub(slashAmount)
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(fmt.Sprintf("Validator (%s) slashed! Delegation updated to: %v", validator.Address, validator.DelegationAmt))

	return nil
}
