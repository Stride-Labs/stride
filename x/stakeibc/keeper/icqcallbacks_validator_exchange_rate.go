package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto" //nolint:staticcheck

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/utils"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args into a Validator struct
	queriedValidator := stakingtypes.Validator{}
	err := k.cdc.Unmarshal(args, &queriedValidator)
	if err != nil {
		return errorsmod.Wrapf(types.ErrMarshalFailure, "unable to unmarshal query response into Validator type, err: %s", err.Error())
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator, "Query response - Validator: %s, Jailed: %v, Tokens: %v, Shares: %v",
		queriedValidator.OperatorAddress, queriedValidator.Jailed, queriedValidator.Tokens, queriedValidator.DelegatorShares))

	// Get the validator from the host zone
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedValidator.OperatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedValidator.OperatorAddress)
	}
	prevTokensToSharesRate := validator.InternalExchangeRate.InternalTokensToSharesRate

	// Get the stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}

	// If the validator's delegation shares is 0, we'll get a division by zero error when trying to get the exchange rate
	//  because `validator.TokensFromShares` uses delegation shares in the denominator
	if queriedValidator.DelegatorShares.IsZero() {
		return errorsmod.Wrapf(types.ErrDivisionByZero,
			"can't calculate validator internal exchange rate because delegation amount is 0 (validator: %s)", validator.Address)
	}

	// We want the validator's internal exchange rate which is held internally behind `validator.TokensFromShares`
	//  Since,
	//     exchange_rate = num_tokens / num_shares
	//  We can use `validator.TokensFromShares`, plug in 1.0 for the number of shares,
	//    and the returned number of tokens will be equal to the internal exchange rate
	currTokensToSharesRate := queriedValidator.TokensFromShares(sdk.NewDec(1.0))
	validator.InternalExchangeRate = &types.ValidatorExchangeRate{
		InternalTokensToSharesRate: currTokensToSharesRate,
		EpochNumber:                strideEpochTracker.EpochNumber,
	}
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator,
		"Previous Validator Exchange Rate: %v, Current Validator Exchange Rate: %v",
		prevTokensToSharesRate, currTokensToSharesRate))

	// Check if the validator was slashed by comparing the exchange rates
	validatorWasSlashed := !prevTokensToSharesRate.IsNil() && prevTokensToSharesRate.Equal(currTokensToSharesRate)

	// Determine if we're in a callback for the LSMLiquidStake by checking if the callback data is non-empty
	// If this query was triggered manually, the callback data will be empty
	inLSMLiquidStakeCallback := len(query.CallbackData) != 0

	// If we are in an LSMLiquidStake callback, unmarshal the callback data
	var lsmLiquidStake types.LSMLiquidStake
	var lsmTokenDeposit recordstypes.LSMTokenDeposit
	if inLSMLiquidStakeCallback {
		var callbackData types.ValidatorExchangeRateQueryCallback
		if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal validator exchange rate callback data")
		}
		lsmLiquidStake = *callbackData.LsmLiquidStake
		lsmTokenDeposit = *lsmLiquidStake.Deposit
	}

	// If the validator was not slashed, and we're NOT in the callback of an LSM Liquid stake,
	// we can stop here - there's no need to query for the delegator shares
	if !validatorWasSlashed && !inLSMLiquidStakeCallback {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator,
			"Validator was not slashed"))
		return nil
	}

	// If the validator was not slashed, and we're in the callback of an LSM liquid stake,
	// finish the transaction to mint the user their stTokens
	if !validatorWasSlashed && inLSMLiquidStakeCallback {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator,
			"Validator was not slashed, finishing LSM liquid stake"))

		if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake); err != nil {
			// If finishing the transaction failed, emit an event, but swallow the error so
			// the query object still gets deleted
			errorMessage := fmt.Sprintf("lsm liquid stake callback failed after slash query: %s", err.Error())
			EmitFailedLSMLiquidStakeEvent(ctx, hostZone, lsmTokenDeposit, errorMessage)
			k.Logger(ctx).Error(errorMessage)
		}
		return nil
	}

	// If we're in the LSM liquid stake callback and there was a slash, reject the transaction by emitting an event
	if inLSMLiquidStakeCallback {
		// The transaction is "rejected", by emitting an event
		errorMessage := fmt.Sprintf("validator was slashed - previous exchange rate: %s, current exchange rate: %s",
			prevTokensToSharesRate, currTokensToSharesRate)

		EmitFailedLSMLiquidStakeEvent(ctx, hostZone, lsmTokenDeposit, errorMessage)
		k.Logger(ctx).Error(errorMessage)
	}

	// If the validator was slashed, we'll have to issue a delegator shares query to determine
	//   the magnitude of the slash
	// If this query was done manually (instead of through an LSM liquid stake)
	//   use a relaxed timeout on this query
	// If this is from an LSM Liquid Stake callback, use an aggressive timeout for the query since
	//   this will block new users from LSM liquid staking
	timeout := uint64(ctx.BlockTime().UnixNano() + (time.Hour).Nanoseconds())
	if inLSMLiquidStakeCallback {
		timeout = uint64(ctx.BlockTime().UnixNano() + (LSMSlashQueryTimeout).Nanoseconds())
	}

	// Now that we're armed with the exchange rate, we can query for the delegator shares and calculated the
	// current delegated tokens
	if err := k.QueryDelegationsIcq(ctx, hostZone, validator, timeout); err != nil {
		return errorsmod.Wrapf(types.ErrICQFailed, "Failed to query delegations, err: %s", err.Error())
	}

	return nil
}
