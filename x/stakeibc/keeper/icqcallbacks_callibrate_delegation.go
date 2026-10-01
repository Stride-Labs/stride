package keeper

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/utils"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// CalibrateDelegationCallback applies an admin correction from queried shares and the
// stored shares-to-tokens rate, only while the query's recorded delegation remains current.
// Stale or unusable snapshots are discarded; admins may submit a fresh calibration.
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

	// Grab the validator object from the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelegation.ValidatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", queriedDelegation.ValidatorAddress)
	}

	// Old calibration queries are purged at v35, so a missing snapshot is never safe to apply.
	var callbackData types.DelegatorSharesQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil || callbackData.InitialValidatorDelegation.IsNil() {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
			"Validator (%s) has missing or invalid calibration snapshot, skipping calibration", validator.Address))
		return nil
	}

	// A completed ICA has already cleared its in-progress counter, but the snapshot still
	// detects its delegation change. Discard the response rather than retrying indefinitely.
	overlapped, err := k.CheckDelegationChangedDuringQuery(ctx, validator, callbackData.InitialValidatorDelegation, validator.Delegation)
	if err != nil {
		return err
	}
	if overlapped {
		return nil
	}

	// Skip if the stored rate is unusable, since the computed token amount would be meaningless
	// and would wipe the recorded delegation
	if validator.SharesToTokensRate.IsNil() || !validator.SharesToTokensRate.IsPositive() {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
			"Validator (%s) has a non-positive shares to tokens rate (%v), skipping calibration",
			validator.Address, validator.SharesToTokensRate))
		return nil
	}

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
