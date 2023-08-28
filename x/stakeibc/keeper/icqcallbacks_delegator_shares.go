package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v14/utils"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// DelegatorSharesCallback is a callback handler for UpdateValidatorSharesExchRate queries.
//
// In an attempt to get the ICA's delegation amount on a given validator, we have to query:
//  1. the validator's internal shares to tokens rate
//  2. the Delegation ICA's delegated shares
//     And apply the following equation:
//     numTokens = numShares * sharesToTokensRate
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
	queriedDelegation := stakingtypes.Delegation{}
	err := k.cdc.Unmarshal(args, &queriedDelegation)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal delegator shares query response into Delegation type")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation, "Query response - Delegator: %s, Validator: %s, Shares: %v",
		queriedDelegation.DelegatorAddress, queriedDelegation.ValidatorAddress, queriedDelegation.Shares))

	// Unmarshal the callback data containing the previous delegation to the validator (from the time the query was submitted)
	var callbackData types.DelegatorSharesQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal delegator shares callback data")
	}

	// Grab the validator object from the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelegation.ValidatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedDelegation.ValidatorAddress)
	}

	// Check if delegation is zero since this will affect measuring the slash amount
	if validator.Delegation.IsZero() {
		return errorsmod.Wrapf(types.ErrNoValidatorAmts, "Current delegation to validator is zero, unable to check slash magnitude %+v", validator)
	}

	// Check if the ICQ overlapped a delegation, undelegation, or detokenization ICA
	// that would have modfied the number of delegated tokens
	prevInternalDelegation := callbackData.InitialValidatorDelegation
	currInternalDelegation := validator.Delegation
	icaOverlappedIcq, err := k.CheckDelegationChangedDuringQuery(ctx, validator, prevInternalDelegation, currInternalDelegation)
	if err != nil {
		return err
	}

	// If the ICA/ICQ overlapped, submit a new query
	if icaOverlappedIcq {
		// Store the updated validator delegation amount
		callbackDataBz, err := proto.Marshal(&types.DelegatorSharesQueryCallback{
			InitialValidatorDelegation: currInternalDelegation,
		})
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal delegator shares callback data")
		}
		query.CallbackData = callbackDataBz

		if err := k.InterchainQueryKeeper.RetryICQRequest(ctx, query); err != nil {
			return errorsmod.Wrapf(err, "unable to resubmit delegator shares query")
		}
		return nil
	}

	// If there was no ICA/ICQ overlap, update the validator to indicate that the query
	//  is no longer in progress (which will unblock LSM liquid stakes to that validator)
	validator.SlashQueryInProgress = false
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	// Confirm the validator was slashed by looking at the number of tokens associated with the delegation
	validatorWasSlashed, delegatedTokens, err := k.CheckForSlash(ctx, hostZone, valIndex, queriedDelegation)
	if err != nil {
		return err
	}
	// If the validator was not slashed, exit now
	if !validatorWasSlashed {
		return nil
	}

	// If the validator was slashed and the query did not overlap any ICAs, update the internal record keeping
	if err := k.SlashValidatorOnHostZone(ctx, hostZone, valIndex, delegatedTokens); err != nil {
		return err
	}

	return nil
}

// The number of tokens returned from the query must be consistent with the tokens
// stored in our internal record keeping during this callback, otherwise the comparision
// between the two is invalidated
//
// As a result, we must avoid a race condition between the ICQ and a delegate, undelegate,
// redelegate, or detokenization ICA
//
// More specifically, we must avoid the following cases:
//
//	Case 1)
//	         ICQ Lands on Host                                          ICQ Ack on Stride
//	                             ICA Lands on Host    ICA Ack on Stride
//	Case 2)
//	         ICA Lands on Host                                          ICA Ack on Stride
//	                             ICQ Lands on Host    ICQ Ack on Stride
//
// We can prevent Case #1 by checking if the delegation total on the validator has changed
// while the query was in flight
//
// We can prevent Case #2 by checking if the validator has a delegation change in progress
func (k Keeper) CheckDelegationChangedDuringQuery(
	ctx sdk.Context,
	validator types.Validator,
	previousInternalDelegation sdkmath.Int,
	currentInternalDelegation sdkmath.Int,
) (overlapped bool, err error) {
	// Confirm the delegation total in the internal record keeping has not changed while the query was inflight
	// If it has changed, exit this callback (to prevent any accounting errors) and resubmit the query
	if !currentInternalDelegation.Equal(previousInternalDelegation) {
		k.Logger(ctx).Error(fmt.Sprintf(
			"Validator (%s) delegation changed while delegator shares query was in flight. Resubmitting query", validator.Address))
		return true, nil
	}

	// Confirm there isn't currently an active delegation change ICA for this validator
	if validator.DelegationChangesInProgress > 0 {
		k.Logger(ctx).Error(fmt.Sprintf(
			"Validator (%s) has %d delegation changing ICAs in progress. Resubmitting query ",
			validator.Address, validator.DelegationChangesInProgress))
		return true, nil
	}

	return false, nil
}

// Check if a slash occured by comparing the validator's sharesToTokens rate and delegator shares
// from the query responses (tokens = shares * sharesToTokensRate)
//
// If the change in delegation only differs by a small precision error, it was likely
// due to an decimal -> int truncation that occurs during unbonding. In this case, still update the validator
//
// If the change in delegation was an increase, the response can't be trusted so an error is thrown
func (k Keeper) CheckForSlash(
	ctx sdk.Context,
	hostZone types.HostZone,
	valIndex int64,
	queriedDelegation stakingtypes.Delegation,
) (validatorWasSlashed bool, delegatedTokens sdkmath.Int, err error) {
	chainId := hostZone.ChainId
	validator := hostZone.Validators[valIndex]

	// Calculate the number of tokens delegated (using the internal sharesToTokensRate)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	delegatedTokens = queriedDelegation.Shares.Mul(validator.SharesToTokensRate).TruncateInt()
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Previous Delegation: %v, Current Delegation: %v", validator.Delegation, delegatedTokens))

	// Confirm the validator has actually been slashed
	if delegatedTokens.Equal(validator.Delegation) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation, "Validator was not slashed"))
		return false, delegatedTokens, nil
	}

	// If the true delegation is slightly higher than our record keeping, this could be due to float imprecision
	// Correct record keeping accordingly
	precisionErrorThreshold := sdkmath.NewInt(25)
	precisionError := delegatedTokens.Sub(validator.Delegation)
	if precisionError.IsPositive() && precisionError.LTE(precisionErrorThreshold) {
		// Update the validator on the host zone
		validator.Delegation = validator.Delegation.Add(precisionError)
		hostZone.TotalDelegations = hostZone.TotalDelegations.Add(precisionError)

		hostZone.Validators[valIndex] = validator
		k.SetHostZone(ctx, hostZone)

		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
			"Delegation updated to %v", validator.Delegation))

		return false, delegatedTokens, nil
	}

	// If the delegation returned from the query is much higher than our record keeping, exit with an error
	if delegatedTokens.GT(validator.Delegation) {
		return false, delegatedTokens, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"Validator (%s) tokens returned from query is greater than the Delegation", validator.Address)
	}

	return true, delegatedTokens, nil
}

// Update the accounting on the host zone and validator to record the slash
// NOTE: we assume any decrease in delegation amt that's not tracked via records is a slash
func (k Keeper) SlashValidatorOnHostZone(ctx sdk.Context, hostZone types.HostZone, valIndex int64, delegatedTokens sdkmath.Int) error {
	chainId := hostZone.ChainId
	validator := hostZone.Validators[valIndex]

	// There is a check upstream to verify that validator.Delegation is not 0
	// This check is to explicitly avoid a division by zero error
	if validator.Delegation.IsZero() {
		return errorsmod.Wrapf(types.ErrDivisionByZero, "Zero Delegation has caused division by zero from validator, %+v", validator)
	}

	// Get slash percentage
	slashAmount := validator.Delegation.Sub(delegatedTokens)
	slashPct := sdk.NewDecFromInt(slashAmount).Quo(sdk.NewDecFromInt(validator.Delegation))
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Validator was slashed! Validator: %s, Delegator: %s, Delegation in State: %v, Delegation from ICQ %v, Slash Amount: %v, Slash Pct: %v",
		validator.Address, hostZone.DelegationIcaAddress, validator.Delegation, delegatedTokens, slashAmount, slashPct))

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
	hostZone.Validators[valIndex] = validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Delegation updated to: %v, Weight updated to: %v", validator.Delegation, validator.Weight))

	// Update the redemption rate
	depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
	k.UpdateRedemptionRateForHostZone(ctx, hostZone, depositRecords)

	return nil
}
