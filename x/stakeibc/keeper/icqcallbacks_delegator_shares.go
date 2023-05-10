package keeper

import (
	sdkmath "cosmossdk.io/math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/utils"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_Delegation,
		"Starting delegator shares callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response which returns a delegation object for the delegator/validator pair
	queriedDelgation := stakingtypes.Delegation{}
	err := k.cdc.Unmarshal(args, &queriedDelgation)
	if err != nil {
		return errorsmod.Wrapf(types.ErrMarshalFailure, "unable to unmarshal query response into Delegation type, err: %s", err.Error())
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation, "Query response - Delegator: %s, Validator: %s, Shares: %v",
		queriedDelgation.DelegatorAddress, queriedDelgation.ValidatorAddress, queriedDelgation.Shares))

	// Ensure ICQ can be issued now, else fail the callback
	withinBufferWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
	}
	if !withinBufferWindow {
		return errorsmod.Wrapf(types.ErrOutsideIcqWindow, "callback is outside ICQ window")
	}

	// Grab the validator object from the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelgation.ValidatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedDelgation.ValidatorAddress)
	}

	// Get the validator's internal exchange rate, aborting if it hasn't been updated this epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "unable to get epoch tracker for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	if validator.InternalExchangeRate.EpochNumber != strideEpochTracker.EpochNumber {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"validator (%s) internal exchange rate has not been updated this epoch (epoch #%d)", validator.Address, strideEpochTracker.EpochNumber)
	}

	// Calculate the number of tokens delegated (using the internal exchange rate)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	delegatedTokens := queriedDelgation.Shares.Mul(validator.InternalExchangeRate.InternalTokensToSharesRate).TruncateInt()
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Previous Delegation: %v, Current Delegation: %v", validator.DelegationAmt, delegatedTokens))

	// Confirm the validator has actually been slashed
	if delegatedTokens.Equal(validator.DelegationAmt) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation, "Validator was not slashed"))
		return nil
	}

	// If the true delegation is slightly higher than our record keeping, this could be due to float imprecision
	// Correct record keeping accordingly
	precisionErrorThreshold := sdkmath.NewInt(25)
	precisionError := delegatedTokens.Sub(validator.DelegationAmt)
	if precisionError.IsPositive() && precisionError.LTE(precisionErrorThreshold) {
		// Update the validator on the host zone
		validator.DelegationAmt = validator.DelegationAmt.Add(precisionError)
		hostZone.StakedBal = hostZone.StakedBal.Add(precisionError)

		hostZone.Validators[valIndex] = &validator
		k.SetHostZone(ctx, hostZone)

		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
			"Delegation updated to %v", validator.DelegationAmt))

		return nil
	}

	// If the delegation returned from the query is much higher than our record keeping, exit with an error
	if delegatedTokens.GT(validator.DelegationAmt) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Validator (%s) tokens returned from query is greater than the DelegationAmt", validator.Address)
	}

	// TODO(TESTS-171) add some safety checks here (e.g. we could query the slashing module to confirm the decr in tokens was due to slash)
	// update our records of the total stakedbal and of the validator's delegation amt
	// NOTE:  we assume any decrease in delegation amt that's not tracked via records is a slash

	// Get slash percentage
	slashAmount := validator.DelegationAmt.Sub(delegatedTokens)
	slashPct := sdk.NewDecFromInt(slashAmount).Quo(sdk.NewDecFromInt(validator.DelegationAmt))
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Validator was slashed! Validator: %s, Delegator: %s, Delegation in State: %v, Delegation from ICQ %v, Slash Amount: %v, Slash Pct: %v",
		validator.Address, queriedDelgation.DelegatorAddress, validator.DelegationAmt, delegatedTokens, slashAmount, slashPct))

	// Abort if the slash was greater than the safety threshold
	slashThreshold, err := cast.ToInt64E(k.GetParam(ctx, types.KeySafetyMaxSlashPercent))
	if err != nil {
		return err
	}
	slashThresholdDecimal := sdk.NewDec(slashThreshold).Quo(sdk.NewDec(100))
	if slashPct.GT(slashThresholdDecimal) {
		return errorsmod.Wrapf(types.ErrSlashExceedsSafetyThreshold,
			"Validator slashed but ABORTING update, slash (%v) is greater than safety threshold (%v)", slashPct, slashThresholdDecimal)
	}

	// Update the validator weight and delegation reflect to reflect the slash
	weight, err := cast.ToInt64E(validator.Weight)
	if err != nil {
		return errorsmod.Wrapf(types.ErrIntCast, "unable to convert validator weight to int64, err: %s", err.Error())
	}
	weightAdjustment := sdk.NewDecFromInt(delegatedTokens).Quo(sdk.NewDecFromInt(validator.DelegationAmt))

	validator.Weight = sdk.NewDec(weight).Mul(weightAdjustment).TruncateInt().Uint64()
	validator.DelegationAmt = validator.DelegationAmt.Sub(slashAmount)

	// Update the validator on the host zone
	hostZone.StakedBal = hostZone.StakedBal.Sub(slashAmount)
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Delegation updated to: %v, Weight updated to: %v", validator.DelegationAmt, validator.Weight))

	return nil
}
