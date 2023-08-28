package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v14/utils"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// ValidatorCallback is a callback handler for validator queries.
//
// In an attempt to get the ICA's delegation amount on a given validator, we have to query:
//  1. the validator's internal sharesToTokens rate
//  2. the Delegation ICA's delegated shares
//     And apply the following equation:
//     numTokens = numShares * sharesToTokensRate
//
// This is the callback from query #1
// We only issue query #2 if the validator sharesToTokens rate from #1 has changed (indicating a slash)
func ValidatorSharesToTokensRateCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_Validator,
		"Starting validator sharesToTokens rate balance callback, QueryId: %vs, QueryType: %s, Connection: %s",
		query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, query.ChainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Determine if we're in a callback for the LSMLiquidStake by checking if the callback data is non-empty
	// If this query was triggered manually, the callback data will be empty
	inLSMLiquidStakeCallback := len(query.CallbackData) != 0

	// If the query timed out, either fail the LSM liquid stake or, if this query was submitted manually, do nothing
	if query.HasTimedOut(ctx.BlockTime()) {
		if inLSMLiquidStakeCallback {
			return k.LSMSlashQueryTimeout(ctx, hostZone, query)
		}
		return nil
	}

	// Unmarshal the query response args into a Validator struct
	queriedValidator := stakingtypes.Validator{}
	if err := k.cdc.Unmarshal(args, &queriedValidator); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal query response into Validator type")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Validator,
		"Query response - Validator: %s, Jailed: %v, Tokens: %v, Shares: %v",
		queriedValidator.OperatorAddress, queriedValidator.Jailed, queriedValidator.Tokens, queriedValidator.DelegatorShares))

	// Check the query response to identify if the validator was slashed
	validatorWasSlashed, err := k.CheckIfValidatorWasSlashed(ctx, hostZone, queriedValidator)
	if err != nil {
		return err
	}

	// If we are in the LSMLiquidStake callback, finish the transaction
	if inLSMLiquidStakeCallback {
		if err := k.LSMSlashQueryCallback(ctx, hostZone, query, validatorWasSlashed); err != nil {
			return errorsmod.Wrapf(err, "unable to finish LSM liquid stake")
		}
	}

	// If the validator was slashed, we'll have to issue a delegator shares query to determine
	// the magnitude of the slash
	if validatorWasSlashed {
		if err := k.SubmitDelegationICQ(ctx, hostZone, queriedValidator.OperatorAddress); err != nil {
			return errorsmod.Wrapf(err, "Failed to submit ICQ validator delegations")
		}
	}

	return nil
}

// Determines if the validator was slashed by comparing the validator sharesToTokens rate from the query response
// with the sharesToTokens rate stored on the validator
func (k Keeper) CheckIfValidatorWasSlashed(
	ctx sdk.Context,
	hostZone types.HostZone,
	queriedValidator stakingtypes.Validator,
) (validatorWasSlashed bool, err error) {
	// Get the validator from the host zone
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedValidator.OperatorAddress)
	if !found {
		return false, errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedValidator.OperatorAddress)
	}
	previousSharesToTokensRate := validator.SharesToTokensRate

	// If the validator's delegation shares is 0, we'll get a division by zero error when trying to get the sharesToTokens rate
	//  because `validator.TokensFromShares` uses delegation shares in the denominator
	if queriedValidator.DelegatorShares.IsZero() {
		return false, errorsmod.Wrapf(types.ErrDivisionByZero,
			"can't calculate validator internal sharesToTokens rate because delegation amount is 0 (validator: %s)", validator.Address)
	}

	// We want the validator's internal sharesToTokens rate which is held internally
	// behind the inverse of the function `validator.TokensFromShares`
	//  Since,
	//     sharesToTokensRate = numTokens / numShares
	//  We can use `validator.TokensFromShares`, plug in 1.0 for the number of shares,
	//    and the returned number of tokens will be equal to the internal sharesToTokens rate
	currentSharesToTokensRate := queriedValidator.TokensFromShares(sdk.NewDec(1.0))
	validator.SharesToTokensRate = currentSharesToTokensRate
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	// Check if the validator was slashed by comparing the sharesToTokens rate from the query
	// with the preivously stored sharesToTokens rate
	previousSharesToTokensRateKnown := !previousSharesToTokensRate.IsNil() && previousSharesToTokensRate.IsPositive()
	validatorWasSlashed = previousSharesToTokensRateKnown && !previousSharesToTokensRate.Equal(currentSharesToTokensRate)

	if !validatorWasSlashed {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(hostZone.ChainId, ICQCallbackID_Validator,
			"Validator was not slashed"))
		return false, nil
	}

	// Emit an event if the validator was slashed
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(hostZone.ChainId, ICQCallbackID_Validator,
		"Previous Validator SharesToTokens Rate: %v, Current Validator SharesToTokens Rate: %v",
		previousSharesToTokensRate, currentSharesToTokensRate))

	EmitValidatorSharesToTokensRateChangeEvent(ctx, hostZone.ChainId, validator.Address, previousSharesToTokensRate, currentSharesToTokensRate)

	return true, nil
}

// Fails the LSM Liquid Stake if the query timed out
func (k Keeper) LSMSlashQueryTimeout(ctx sdk.Context, hostZone types.HostZone, query icqtypes.Query) error {
	var callbackData types.ValidatorSharesToTokensQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal validator sharesToTokens rate callback data")
	}
	lsmLiquidStake := *callbackData.LsmLiquidStake

	k.FailLSMLiquidStake(ctx, hostZone, lsmLiquidStake, "query timed out")
	return nil
}

// Callback handler for if the slash query was initiated by an LSMLiquidStake transaction
// If the validator was slashed, the LSMLiquidStake should be rejected
// If the validator was not slashed, the LSMLiquidStake should finish to mint the user stTokens
func (k Keeper) LSMSlashQueryCallback(
	ctx sdk.Context,
	hostZone types.HostZone,
	query icqtypes.Query,
	validatorWasSlashed bool,
) error {
	var callbackData types.ValidatorSharesToTokensQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal validator sharesToTokens rate callback data")
	}
	lsmLiquidStake := *callbackData.LsmLiquidStake

	// If the validator was slashed, fail the liquid stake
	if validatorWasSlashed {
		k.FailLSMLiquidStake(ctx, hostZone, lsmLiquidStake, "validator was slashed, failing LSMLiquidStake")
		return nil
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(hostZone.ChainId, ICQCallbackID_Validator,
		"Validator was not slashed, finishing LSM liquid stake"))

	// Finish the LSMLiquidStake with a temporary context so that the state changes can
	// be discarded if it errors
	err := utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		async := true
		return k.FinishLSMLiquidStake(ctx, lsmLiquidStake, async)
	})

	// If finishing the transaction failed, emit an event and remove the LSMTokenDeposit
	if err != nil {
		k.FailLSMLiquidStake(ctx, hostZone, lsmLiquidStake,
			fmt.Sprintf("lsm liquid stake callback failed after slash query: %s", err.Error()))
	}

	return nil
}

// Fail an LSMLiquidStake transaction by emitting an event and removing the LSMTokenDeposit record
func (k Keeper) FailLSMLiquidStake(ctx sdk.Context, hostZone types.HostZone, lsmLiquidStake types.LSMLiquidStake, errorMessage string) {
	EmitFailedLSMLiquidStakeEvent(ctx, hostZone, *lsmLiquidStake.Deposit, errorMessage)
	k.Logger(ctx).Error(errorMessage)

	// Remove the LSMTokenDeposit
	k.RecordsKeeper.RemoveLSMTokenDeposit(ctx, lsmLiquidStake.Deposit.ChainId, lsmLiquidStake.Deposit.Denom)
}
