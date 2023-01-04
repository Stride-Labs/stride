package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v4/utils"
	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// ValidatorCallback is a callback handler for validator queries.
//
// In an attempt to get the ICA's delegation amount on a given validator, we have to query:
//  1. the validator's internal exchange rate
//  2. the Delegation ICA's delegated shares
//     And apply the following equation:
//     num_tokens = exchange_rate * num_shares
//
// This is the callback from query #1
func ValidatorExchangeRateCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_Validator,
		"Starting validator exchange rate balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, query.ChainId)
	if !found {
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args into a Validator struct
	queriedValidator := stakingtypes.Validator{}
	err := k.cdc.Unmarshal(args, &queriedValidator)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrMarshalFailure, "unable to unmarshal query response into Validator type, err: %s", err.Error())
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator, "Query response - Validator: %s, Jailed: %v, Tokens: %v, Shares: %v",
		queriedValidator.OperatorAddress, queriedValidator.Jailed, queriedValidator.Tokens, queriedValidator.DelegatorShares))

	// Ensure ICQ can be issued now, else fail the callback
	withinBufferWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
	}
	if !withinBufferWindow {
		// QUESTION: Should this return an error?
		return sdkerrors.Wrapf(types.ErrOutsideIcqWindow, "callback is outside ICQ window")
	}

	// Get the validator from the host zone
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedValidator.OperatorAddress)
	if !found {
		return sdkerrors.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedValidator.OperatorAddress)
	}

	// Get the stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}

	// If the validator's delegation shares is 0, we'll get a division by zero error when trying to get the exchange rate
	//  because `validator.TokensFromShares` uses delegation shares in the denominator
	if queriedValidator.DelegatorShares.IsZero() {
		return sdkerrors.Wrapf(types.ErrDivisionByZero,
			"can't calculate validator internal exchange rate because delegation amount is 0 (validator: %s)", validator.Address)
	}

	// We want the validator's internal exchange rate which is held internally behind `validator.TokensFromShares`
	//  Since,
	//     exchange_rate = num_tokens / num_shares
	//  We can use `validator.TokensFromShares`, plug in 1.0 for the number of shares,
	//    and the returned number of tokens will be equal to the internal exchange rate
	validator.InternalExchangeRate = &types.ValidatorExchangeRate{
		InternalTokensToSharesRate: queriedValidator.TokensFromShares(sdk.NewDec(1.0)),
		EpochNumber:                strideEpochTracker.GetEpochNumber(),
	}
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator, "Validator Internal Exchange Rate: %v",
		validator.InternalExchangeRate.InternalTokensToSharesRate))

	// Armed with the exch rate, we can now query the (validator,delegator) delegation
	if err := k.QueryDelegationsIcq(ctx, hostZone, queriedValidator.OperatorAddress); err != nil {
		return sdkerrors.Wrapf(types.ErrICQFailed, "Failed to query delegations")
	}

	return nil
}
