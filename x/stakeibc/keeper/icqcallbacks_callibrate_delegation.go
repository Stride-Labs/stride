package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/utils"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
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

	// An empty response means the delegation ICA has no delegation to the validator on the host
	// (the query opts into reaching this callback empty): the shares are zero and the validator
	// comes from the callback data, since there is no Delegation to read it from
	isEmptyResponse := len(args) == 0
	queriedDelegation := stakingtypes.Delegation{Shares: sdkmath.LegacyZeroDec()}
	if isEmptyResponse {
		if len(query.CallbackData) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "empty calibration response without a validator address in the query callback data")
		}
		queriedDelegation.ValidatorAddress = string(query.CallbackData)
	} else {
		// Unmarshal the query response which returns a delegation object for the delegator/validator pair
		if err := k.cdc.Unmarshal(args, &queriedDelegation); err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal delegator shares query response into Delegation type")
		}
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate, "Query response - Delegator: %s, Validator: %s, Shares: %v",
		queriedDelegation.DelegatorAddress, queriedDelegation.ValidatorAddress, queriedDelegation.Shares))

	// Grab the validator object from the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelegation.ValidatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedDelegation.ValidatorAddress)
	}

	// Skip if there is an active delegation change ICA for this validator, since the queried
	// shares race the recorded delegation
	if validator.DelegationChangesInProgress > 0 {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
			"Validator (%s) has %d delegation changing ICAs in progress, skipping calibration",
			validator.Address, validator.DelegationChangesInProgress))
		return nil
	}

	// Skip if the stored rate is unusable, since the computed token amount would be meaningless
	// and would wipe the recorded delegation (not needed for an empty response: zero shares are zero tokens at any rate)
	if !isEmptyResponse && (validator.SharesToTokensRate.IsNil() || !validator.SharesToTokensRate.IsPositive()) {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
			"Validator (%s) has a non-positive shares to tokens rate (%v), skipping calibration",
			validator.Address, validator.SharesToTokensRate))
		return nil
	}

	// Calculate the number of tokens delegated (using the internal sharesToTokensRate)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	delegatedTokens := sdkmath.ZeroInt()
	if !isEmptyResponse {
		delegatedTokens = queriedDelegation.Shares.Mul(validator.SharesToTokensRate).TruncateInt()
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
		"Previous Delegation: %v, Current Delegation: %v", validator.Delegation, delegatedTokens))

	// Confirm the validator has actually been slashed
	if delegatedTokens.Equal(validator.Delegation) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate, "Validator delegation is correct"))
		return nil
	}

	// Apply the whole difference. The 5,000 base-unit cap that used to bound this existed to
	// limit what a permissionless caller could move; MsgCalibrateDelegation is admin-only now
	// (wind-down spec §5) and the day-0 refresh needs to true up drifts of any size
	// Note: There should be no stateful changes above this line
	delegationChange := validator.Delegation.Sub(delegatedTokens)
	validator.Delegation = validator.Delegation.Sub(delegationChange)
	hostZone.TotalDelegations = hostZone.TotalDelegations.Sub(delegationChange)

	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
		"Delegation updated to: %v", validator.Delegation))

	return nil
}
