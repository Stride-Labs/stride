package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v5/utils"
	epochtypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
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
		return fmt.Errorf("no registered zone for queried chain ID (%s): %s", chainId, types.ErrHostZoneNotFound.Error())
	}

	// Unmarshal the query response args into a Validator struct
	queriedValidator := stakingtypes.Validator{}
	err := k.cdc.Unmarshal(args, &queriedValidator)
	if err != nil {
		return fmt.Errorf("unable to unmarshal query response into Validator type, err: %s: %s", err.Error(), types.ErrMarshalFailure.Error())
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator, "Query response - Validator: %s, Jailed: %v, Tokens: %v, Shares: %v",
		queriedValidator.OperatorAddress, queriedValidator.Jailed, queriedValidator.Tokens, queriedValidator.DelegatorShares))

	// Ensure ICQ can be issued now, else fail the callback
	withinBufferWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine if ICQ callback is inside buffer window, err: %s: %s", err.Error(), types.ErrInvalidRequest.Error())
	}
	if !withinBufferWindow {
		return fmt.Errorf("callback is outside ICQ window: %s", types.ErrOutsideIcqWindow.Error())
	}

	// Get the validator from the host zone
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedValidator.OperatorAddress)
	if !found {
		return fmt.Errorf("no registered validator for address (%s): %s", queriedValidator.OperatorAddress, types.ErrValidatorNotFound.Error())
	}

	// Get the stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return fmt.Errorf("no epoch number for epoch (%s): %s", epochtypes.STRIDE_EPOCH, types.ErrNotFound.Error())
	}

	// If the validator's delegation shares is 0, we'll get a division by zero error when trying to get the exchange rate
	//  because `validator.TokensFromShares` uses delegation shares in the denominator
	if queriedValidator.DelegatorShares.IsZero() {
		return fmt.Errorf("can't calculate validator internal exchange rate because delegation amount is 0 (validator: %s): %s", validator.Address, types.ErrDivisionByZero.Error())
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
		return fmt.Errorf("Failed to query delegations, err: %s: %s", err.Error(), types.ErrICQFailed.Error())
	}

	return nil
}
