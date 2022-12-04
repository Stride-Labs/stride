package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	hostZone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		errMsg := fmt.Sprintf("no registered zone for queried chain ID (%s)", query.GetChainId())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}
	queriedValidator := stakingtypes.Validator{}
	err := k.cdc.Unmarshal(args, &queriedValidator)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal queriedValidator info for zone %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("ValidatorCallback: HostZone %s, Queried Validator %v, Jailed: %v, Tokens: %v, Shares: %v",
		hostZone.ChainId, queriedValidator.OperatorAddress, queriedValidator.Jailed, queriedValidator.Tokens, queriedValidator.DelegatorShares))

	// ensure ICQ can be issued now! else fail the callback
	withinBufferWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrOutsideIcqWindow, errMsg)
	} else if !withinBufferWindow {
		k.Logger(ctx).Error("validator exchange rate callback is outside ICQ window")
		return nil
	}

	// get the validator from the host zone
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedValidator.OperatorAddress)
	if !found {
		errMsg := fmt.Sprintf("no registered validator for address (%s)", queriedValidator.OperatorAddress)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrValidatorNotFound, errMsg)
	}
	// get the stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}

	// If the validator's delegation shares is 0, we'll get a division by zero error when trying to get the exchange rate
	//  because `validator.TokensFromShares` uses delegation shares in the denominator
	if queriedValidator.DelegatorShares.IsZero() {
		errMsg := fmt.Sprintf("can't calculate validator internal exchange rate because delegation amount is 0 (validator: %s)", validator.Address)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrDivisionByZero, errMsg)
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

	k.Logger(ctx).Info(fmt.Sprintf("ValidatorCallback: HostZone %s, Validator %v, tokensFromShares %v",
		hostZone.ChainId, validator.Address, validator.InternalExchangeRate.InternalTokensToSharesRate))

	// armed with the exch rate, we can now query the (val,del) delegation
	err = k.QueryDelegationsIcq(ctx, hostZone, queriedValidator.OperatorAddress)
	if err != nil {
		errMsg := fmt.Sprintf("ValidatorCallback: failed to query delegation, zone %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICQFailed, errMsg)
	}
	return nil
}
