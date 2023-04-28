package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto" //nolint:staticcheck
	"github.com/spf13/cast"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/utils"
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
		return errorsmod.Wrapf(err, "unable to unmarshal delegator shares query response into Delegation type")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation, "Query response - Delegator: %s, Validator: %s, Shares: %v",
		queriedDelgation.DelegatorAddress, queriedDelgation.ValidatorAddress, queriedDelgation.Shares))

	// Unmarshal the callback data containing the previous delegation to the validator (from the time the query was submitted)
	var callbackData types.DelegatorSharesQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal delegator shares callback data")
	}

	// Grab the validator object from the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelgation.ValidatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedDelgation.ValidatorAddress)
	}

	// Confirm the delegation total in the internal record keeping has not changed while the query was inflight
	// If it has changed, exit this callback (to prevent any accounting errors) and resubmit the query
	if !validator.Delegation.Equal(callbackData.InitialValidatorDelegation) {
		k.Logger(ctx).Error(fmt.Sprintf("Validator (%s) delegation changed while delegator shares query was in flight. Resubmitting query", chainId))

		// Reset the query timeout and resubmit
		query.Timeout = uint64(ctx.BlockTime().UnixNano() + (callbackData.TimeoutDuration).Nanoseconds())
		if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, true); err != nil {
			return errorsmod.Wrapf(err, "unable to resubmit delegator shares query")
		}

		return nil
	}

	// Calculate the number of tokens delegated (using the internal exchange rate)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	delegatedTokens := queriedDelgation.Shares.Mul(validator.InternalShareToTokensRate).TruncateInt()
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Previous Delegation: %v, Current Delegation: %v", validator.Delegation, delegatedTokens))

	// Confirm the validator has actually been slashed
	if delegatedTokens.Equal(validator.Delegation) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation, "Validator was not slashed"))
		return nil
	}

	// If the true delegation is slightly higher than our record keeping, this could be due to float imprecision
	// Correct record keeping accordingly
	precisionErrorThreshold := sdkmath.NewInt(25)
	precisionError := delegatedTokens.Sub(validator.Delegation)
	if precisionError.IsPositive() && precisionError.LTE(precisionErrorThreshold) {
		// Update the validator on the host zone
		validator.Delegation = validator.Delegation.Add(precisionError)
		hostZone.TotalDelegations = hostZone.TotalDelegations.Add(precisionError)

		hostZone.Validators[valIndex] = &validator
		k.SetHostZone(ctx, hostZone)

		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
			"Delegation updated to %v", validator.Delegation))

		return nil
	}

	// If the delegation returned from the query is much higher than our record keeping, exit with an error
	if delegatedTokens.GT(validator.Delegation) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Validator (%s) tokens returned from query is greater than the Delegation", validator.Address)
	}

	// TODO(TESTS-171) add some safety checks here (e.g. we could query the slashing module to confirm the decr in tokens was due to slash)
	// update our records of the total delegation and of the validator's delegation
	// NOTE:  we assume any decrease in delegation amt that's not tracked via records is a slash

	// Get slash percentage
	slashAmount := validator.Delegation.Sub(delegatedTokens)
	slashPct := sdk.NewDecFromInt(slashAmount).Quo(sdk.NewDecFromInt(validator.Delegation))
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Validator was slashed! Validator: %s, Delegator: %s, Delegation in State: %v, Delegation from ICQ %v, Slash Amount: %v, Slash Pct: %v",
		validator.Address, queriedDelgation.DelegatorAddress, validator.Delegation, delegatedTokens, slashAmount, slashPct))

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
	weightAdjustment := sdk.NewDecFromInt(delegatedTokens).Quo(sdk.NewDecFromInt(validator.Delegation))

	validator.Weight = sdk.NewDec(weight).Mul(weightAdjustment).TruncateInt().Uint64()
	validator.Delegation = validator.Delegation.Sub(slashAmount)

	// Update the validator on the host zone
	hostZone.TotalDelegations = hostZone.TotalDelegations.Sub(slashAmount)
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Delegation updated to: %v, Weight updated to: %v", validator.Delegation, validator.Weight))

	return nil
}
