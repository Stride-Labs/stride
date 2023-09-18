package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v15/utils"
	icqtypes "github.com/Stride-Labs/stride/v15/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v15/x/stakeibc/types"
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
func CalibrateDelegationCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_Calibrate,
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
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate, "Query response - Delegator: %s, Validator: %s, Shares: %v",
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

	// Calculate the number of tokens delegated (using the internal sharesToTokensRate)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	delegatedTokens := queriedDelegation.Shares.Mul(validator.SharesToTokensRate).TruncateInt()
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
		"Previous Delegation: %v, Current Delegation: %v", validator.Delegation, delegatedTokens))

	// Confirm the validator has actually been slashed
	if delegatedTokens.Equal(validator.Delegation) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate, "Validator delegation is correct"))
		return nil
	}

	delegationChange := validator.Delegation.Sub(delegatedTokens)
	validator.Delegation = validator.Delegation.Sub(delegationChange)
	hostZone.TotalDelegations = hostZone.TotalDelegations.Sub(delegationChange)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
		"Delegation updated to: %v", validator.Delegation))

	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	return nil
}
