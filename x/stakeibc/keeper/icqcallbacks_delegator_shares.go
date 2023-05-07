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
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
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

	// Check if the ICQ overlapped a delegation, undelegation, or detokenization ICA
	// that would have modfied the number of delegated tokens
	prevInternalDelegation := callbackData.InitialValidatorDelegation
	currInternalDelegation := validator.Delegation
	icaOverlappedIcq, err := k.CheckDelegationChangedDuringQuery(
		ctx,
		chainId,
		validator.Address,
		prevInternalDelegation,
		currInternalDelegation,
	)
	if err != nil {
		return err
	}

	// If the ICA/ICQ overlapped, submit a new query
	if icaOverlappedIcq {
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
//   stored in our internal record keeping during this callback, otherwise the comparision
//   between the two is invalidated
// As a result, we must avoid a race condition between the ICQ and an delegate or undelegate ICA
//
// More specifically, we must avoid the following cases:
//  Case 1)
//           ICQ Lands on Host                                          ICQ Ack on Stride
//                               ICA Lands on Host    ICA Ack on Stride
//  Case 2)
//           ICA Lands on Host                                          ICA Ack on Stride
//                               ICQ Lands on Host    ICQ Ack on Stride
//
// We can prevent Case #1 by checking if the delegation total on the validator has changed
//   while the query was in flight
// We can prevent Case #2 by checking if there are any delegation unbonding records
//   in state IN_PROGRESS (meaning an ICA is in flight)
func (k Keeper) CheckDelegationChangedDuringQuery(
	ctx sdk.Context,
	chainId string,
	validatorAddress string,
	previousInternalDelegation sdkmath.Int,
	currentInternalDelegation sdkmath.Int,
) (overlapped bool, err error) {
	// Confirm the delegation total in the internal record keeping has not changed while the query was inflight
	// If it has changed, exit this callback (to prevent any accounting errors) and resubmit the query
	if !currentInternalDelegation.Equal(previousInternalDelegation) {
		k.Logger(ctx).Error(fmt.Sprintf(
			"Validator (%s) delegation changed while delegator shares query was in flight. Resubmitting query", chainId))
		return true, nil
	}

	// Check that there are no deposit records in state IN_PROGRESS - indicative of a Delegation ICA
	for _, depositRecord := range k.RecordsKeeper.GetAllDepositRecord(ctx) {
		if depositRecord.HostZoneId == chainId &&
			depositRecord.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS {
			k.Logger(ctx).Error("Delegation ICA is currently in progress. Rejecting query callback and resubmitting query")
			return true, nil
		}
	}

	// Check that there are no epoch unbonding records in state IN_PROGRESS - indicative of an Undelegation ICA
	// TODO: This is expensive, we should store these more efficiently
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		for _, hostUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			if hostUnbondingRecord.HostZoneId == chainId &&
				hostUnbondingRecord.Status == recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
				k.Logger(ctx).Error("Undelegation ICA is currently in progress. Rejecting query callback and resubmitting query")
				return true, nil
			}
		}
	}

	// Check that there are no LSMTokenDeposits in state IN_PROGRESS - indictive of an Detokenization ICA
	// We also check for transfer in progress or detokenizaiotn
	for _, lsmTokenDeposit := range k.RecordsKeeper.GetLSMDepositsForHostZone(ctx, chainId) {
		if lsmTokenDeposit.ValidatorAddress == validatorAddress &&
			lsmTokenDeposit.Status == recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS {
			k.Logger(ctx).Error("Detokenization ICA is currently in progress. Rejecting query callback and resubmitting query")
			return true, nil
		}
	}

	return false, nil
}

// Check if a slash occured by comparing the validator's exchange rate  and delegator shares
//   from the query responses (tokens = exchange rate * shares)
// If the change in delegation only differs by a small precision error, it was likely
//   due to an decimal -> int truncation that occurs during unbonding. In this case, still update the validator
// If the change in delegation was an increase, the response can't be trusted so an error is thrown
func (k Keeper) CheckForSlash(
	ctx sdk.Context,
	hostZone types.HostZone,
	valIndex int64,
	queriedDelegation stakingtypes.Delegation,
) (validatorWasSlashed bool, delegatedTokens sdkmath.Int, err error) {
	chainId := hostZone.ChainId
	validator := hostZone.Validators[valIndex]

	// Calculate the number of tokens delegated (using the internal exchange rate)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	delegatedTokens = queriedDelegation.Shares.Mul(validator.InternalSharesToTokensRate).TruncateInt()
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

	// Get slash percentage
	slashAmount := validator.Delegation.Sub(delegatedTokens)
	slashPct := sdk.NewDecFromInt(slashAmount).Quo(sdk.NewDecFromInt(validator.Delegation))
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Validator was slashed! Validator: %s, Delegator: %s, Delegation in State: %v, Delegation from ICQ %v, Slash Amount: %v, Slash Pct: %v",
		validator.Address, hostZone.DelegationIcaAddress, validator.Delegation, delegatedTokens, slashAmount, slashPct))

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
	hostZone.Validators[valIndex] = validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Delegation updated to: %v, Weight updated to: %v", validator.Delegation, validator.Weight))

	return nil
}
