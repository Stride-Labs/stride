package keeper

import (
	"errors"
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v27/utils"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

type ValidatorUnbondCapacity struct {
	ValidatorAddress   string
	CurrentDelegation  sdkmath.Int
	BalancedDelegation sdkmath.Int
	Capacity           sdkmath.Int
}

// The ratio of ideal balanced delegation to the current delegation
// This represents how proportionally unbalanced each validator is
// The smaller number means their current delegation is much larger
// then their fair portion of the current total stake
func (c *ValidatorUnbondCapacity) GetBalanceRatio() (sdk.Dec, error) {
	// ValidatorUnbondCapaciy structs only exist for validators with positive capacity
	//   capacity is CurrentDelegation - BalancedDelegation
	//   positive capacity means CurrentDelegation must be >0

	// Therefore the current delegation here should never be zero
	if c.CurrentDelegation.IsZero() {
		errMsg := fmt.Sprintf("CurrentDelegation should not be 0 inside GetBalanceRatio(), %+v", c)
		return sdk.ZeroDec(), errors.New(errMsg)
	}
	return sdk.NewDecFromInt(c.BalancedDelegation).Quo(sdk.NewDecFromInt(c.CurrentDelegation)), nil
}

// Returns all the host zone unbonding records that should unbond this epoch
// Records are returned as a mapping of epoch unbonding record ID to host zone unbonding record
// Records ready to be unbonded are identified by status UNBONDING_QUEUE or UNBONDING_RETRY_QUEUE
// and a non-zero native amount
func (k Keeper) GetQueuedHostZoneUnbondingRecords(
	ctx sdk.Context,
	chainId string,
) (epochNumbers []uint64, epochToHostZoneUnbondingMap map[uint64]recordstypes.HostZoneUnbonding) {
	epochToHostZoneUnbondingMap = map[uint64]recordstypes.HostZoneUnbonding{}
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbonding.EpochNumber, chainId)
		if !found {
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Epoch %d - Status: %s, Amount: %v",
			epochUnbonding.EpochNumber, hostZoneRecord.Status, hostZoneRecord.NativeTokenAmount))

		if hostZoneRecord.ShouldInitiateUnbonding() || hostZoneRecord.ShouldRetryUnbonding() {
			epochNumbers = append(epochNumbers, epochUnbonding.EpochNumber)
			epochToHostZoneUnbondingMap[epochUnbonding.EpochNumber] = *hostZoneRecord
		}
	}
	return epochNumbers, epochToHostZoneUnbondingMap
}

// Loops through each epoch unbonding record and returns the total number of native tokens
// that should be unbonded
func (k Keeper) GetTotalUnbondAmount(hostZoneUnbondingRecords map[uint64]recordstypes.HostZoneUnbonding) sdkmath.Int {
	totalNativeAmount := sdk.ZeroInt()
	for _, hostZoneRecord := range hostZoneUnbondingRecords {
		totalNativeAmount = totalNativeAmount.Add(hostZoneRecord.NativeTokensToUnbond)
	}
	return totalNativeAmount
}

// Given a list of user redemption record IDs and a redemption rate, sets the native token
// amount on each record, calculated from the stAmount and redemption rate, and returns the
// sum of all native token amounts across all user redemption records
func (k Keeper) RefreshUserRedemptionRecordNativeAmounts(
	ctx sdk.Context,
	chainId string,
	userRedemptionRecordIds []string,
	redemptionRate sdk.Dec,
) (totalNativeAmount sdkmath.Int) {
	// Loop and set the native amount for each record, keeping track of the total
	totalNativeAmount = sdkmath.ZeroInt()
	for _, userRedemptionRecordId := range userRedemptionRecordIds {
		userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
		if !found {
			k.Logger(ctx).Error(utils.LogWithHostZone(chainId, "No user redemption record found for id %s", userRedemptionRecordId))
			continue
		}

		// Calculate the number of native tokens using the redemption rate
		nativeAmount := sdk.NewDecFromInt(userRedemptionRecord.StTokenAmount).Mul(redemptionRate).TruncateInt()
		totalNativeAmount = totalNativeAmount.Add(nativeAmount)

		// Set the native amount on the record
		userRedemptionRecord.NativeTokenAmount = nativeAmount
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
	}
	return totalNativeAmount
}

// Sets the native token amount unbonded on the host zone unbonding record and the associated user redemption records
func (k Keeper) RefreshHostZoneUnbondingNativeTokenAmount(
	ctx sdk.Context,
	epochNumber uint64,
	hostZoneUnbondingRecord recordstypes.HostZoneUnbonding,
) error {
	// Grab the redemption rate from the host zone (to use in the native token calculation)
	chainId := hostZoneUnbondingRecord.HostZoneId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone %s not found", chainId)
	}

	// Set all native token amount on each user redemption record
	redemptionRecordIds := hostZoneUnbondingRecord.UserRedemptionRecords
	totalNativeAmount := k.RefreshUserRedemptionRecordNativeAmounts(ctx, chainId, redemptionRecordIds, hostZone.RedemptionRate)

	// Then set the total on the host zone unbonding record
	hostZoneUnbondingRecord.NativeTokenAmount = totalNativeAmount
	hostZoneUnbondingRecord.NativeTokensToUnbond = totalNativeAmount
	hostZoneUnbondingRecord.StTokensToBurn = hostZoneUnbondingRecord.StTokenAmount
	return k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, chainId, hostZoneUnbondingRecord)
}

// Given a mapping of epoch unbonding record IDs to host zone unbonding records,
// sets the native token amount across all epoch unbonding records, host zone unbonding records,
// and user redemption records, using the most updated redemption rate
// Returns an updated mapping with the refreshed amounts
func (k Keeper) RefreshUnbondingNativeTokenAmounts(
	ctx sdk.Context,
	chainId string,
	hostZoneUnbondings map[uint64]recordstypes.HostZoneUnbonding,
) (refreshedHostZoneUnbondings map[uint64]recordstypes.HostZoneUnbonding, err error) {
	// Refresh the amount for all records in status UNBONDING_QUEUE
	// We don't want to refresh the failed unbonding records
	for _, epochNumber := range utils.Uint64MapKeys(hostZoneUnbondings) {
		hostZoneUnbondingRecord := hostZoneUnbondings[epochNumber]
		if hostZoneUnbondingRecord.Status != recordstypes.HostZoneUnbonding_UNBONDING_QUEUE {
			continue
		}
		if err := k.RefreshHostZoneUnbondingNativeTokenAmount(ctx, epochNumber, hostZoneUnbondingRecord); err != nil {
			return nil, err
		}
	}

	// Once the values have been refreshed, return an updated map
	_, refreshedHostZoneUnbondings = k.GetQueuedHostZoneUnbondingRecords(ctx, chainId)

	return refreshedHostZoneUnbondings, nil
}

// Determine the unbonding capacity that each validator has
// The capacity is determined by the difference between their current delegation
// and their fair portion of the total stake based on their weights
// (i.e. their balanced delegation)
//
// Validators with a balanced delegation less than their current delegation
// are already at a deficit, are not included in the returned list,
// and thus, will not incur any unbonding
func (k Keeper) GetValidatorUnbondCapacity(
	ctx sdk.Context,
	validators []*types.Validator,
	balancedDelegation map[string]sdkmath.Int,
) (validatorCapacities []ValidatorUnbondCapacity) {
	for _, validator := range validators {
		// The capacity equals the difference between their current delegation and
		// the balanced delegation
		// If the capacity is negative, that means the validator has less than their
		// balanced portion. Ignore this case so they don't unbond anything
		balancedDelegation, ok := balancedDelegation[validator.Address]
		if !ok {
			continue
		}

		capacity := validator.Delegation.Sub(balancedDelegation)
		if capacity.IsPositive() {
			validatorCapacities = append(validatorCapacities, ValidatorUnbondCapacity{
				ValidatorAddress:   validator.Address,
				Capacity:           capacity,
				CurrentDelegation:  validator.Delegation,
				BalancedDelegation: balancedDelegation,
			})
		}
	}

	return validatorCapacities
}

// Sort validators by the ratio of the ideal balanced delegation to their current delegation
// This will sort the validator's by how proportionally unbalanced they are
//
// Ex:
//
//	Val1: Ideal Balanced Delegation 80,  Current Delegation 100 (surplus of 20), Ratio: 0.8
//	Val2: Ideal Balanced Delegation 480, Current Delegation 500 (surplus of 20), Ratio: 0.96
//
// While both validators have the same net unbalanced delegation, Val2 is proportionally
// more balanced since the surplus is a smaller percentage of it's overall delegation
//
// This will also sort such that 0-weight validator's will come first as their
// ideal balanced delegation will always be 0, and thus their ratio will always be 0
// If the ratio's are equal, the validator with the larger delegation/capacity will come first
func SortUnbondingCapacityByPriority(validatorUnbondCapacity []ValidatorUnbondCapacity) ([]ValidatorUnbondCapacity, error) {
	// Loop through all validators to make sure none error when getting the balance ratio needed for sorting
	for _, validator := range validatorUnbondCapacity {
		if _, err := validator.GetBalanceRatio(); err != nil {
			return nil, err
		}
	}

	// Pairwise-compare function for Slice Stable Sort
	lessFunc := func(i, j int) bool {
		validatorA := validatorUnbondCapacity[i]
		validatorB := validatorUnbondCapacity[j]

		balanceRatioValA, _ := validatorA.GetBalanceRatio()
		balanceRatioValB, _ := validatorB.GetBalanceRatio()

		// Sort by the balance ratio first - in ascending order - so the more unbalanced validators appear first
		if !balanceRatioValA.Equal(balanceRatioValB) {
			return balanceRatioValA.LT(balanceRatioValB)
		}

		// If the ratio's are equal, use the capacity as a tie breaker
		// where the larget capacity comes first
		if !validatorA.Capacity.Equal(validatorB.Capacity) {
			return validatorA.Capacity.GT(validatorB.Capacity)
		}

		// Finally, if the ratio and capacity are both equal, use address as a tie breaker
		return validatorA.ValidatorAddress < validatorB.ValidatorAddress
	}
	sort.SliceStable(validatorUnbondCapacity, lessFunc)

	return validatorUnbondCapacity, nil
}

// Given a total unbond amount and list of unbond capacity for each validator, sorted by unbond priority
// Iterates through the list and unbonds as much as possible from each validator until all the
// unbonding has been accounted for
//
// Returns the list of messages and the callback data for the ICA
func (k Keeper) GetUnbondingICAMessages(
	hostZone types.HostZone,
	totalUnbondAmount sdkmath.Int,
	prioritizedUnbondCapacity []ValidatorUnbondCapacity,
) (msgs []proto.Message, unbondings []*types.SplitUndelegation, err error) {
	// Loop through each validator and unbond as much as possible
	remainingUnbondAmount := totalUnbondAmount
	for _, validatorCapacity := range prioritizedUnbondCapacity {
		// Break once all unbonding has been accounted for
		if remainingUnbondAmount.IsZero() {
			break
		}

		// Unbond either up to the capacity or up to the total remaining unbond amount
		// (whichever comes first)
		var unbondAmount sdkmath.Int
		if validatorCapacity.Capacity.LT(remainingUnbondAmount) {
			unbondAmount = validatorCapacity.Capacity
		} else {
			unbondAmount = remainingUnbondAmount
		}
		remainingUnbondAmount = remainingUnbondAmount.Sub(unbondAmount)

		// Build the validator splits for the callback
		unbondings = append(unbondings, &types.SplitUndelegation{
			Validator:         validatorCapacity.ValidatorAddress,
			NativeTokenAmount: unbondAmount,
		})
	}

	// Build the undelegate ICA messages from the splits
	for _, unbonding := range unbondings {
		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: unbonding.Validator,
			Amount:           sdk.NewCoin(hostZone.HostDenom, unbonding.NativeTokenAmount),
		})
	}

	// Sanity check that we had enough capacity to unbond
	if !remainingUnbondAmount.IsZero() {
		return msgs, unbondings,
			fmt.Errorf("unable to unbond full amount (%v) from %v", totalUnbondAmount, hostZone.ChainId)
	}

	return msgs, unbondings, nil
}

// Submit undelegate ICA messages in small batches to reduce the gas size per tx
func (k Keeper) BatchSubmitUndelegateICAMessages(
	ctx sdk.Context,
	hostZone types.HostZone,
	epochUnbondingRecordIds []uint64,
	msgs []proto.Message,
	unbondings []*types.SplitUndelegation,
	batchSize int,
) (numTxsSubmitted uint64, err error) {
	// Iterate the full list of messages and submit in batches
	for start := 0; start < len(msgs); start += batchSize {
		end := start + batchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		msgsBatch := msgs[start:end]
		unbondingsBatch := unbondings[start:end]

		// Store the callback data
		undelegateCallback := types.UndelegateCallback{
			HostZoneId:              hostZone.ChainId,
			SplitUndelegations:      unbondingsBatch,
			EpochUnbondingRecordIds: epochUnbondingRecordIds,
		}
		callbackArgsBz, err := proto.Marshal(&undelegateCallback)
		if err != nil {
			return numTxsSubmitted, errorsmod.Wrap(err, "unable to marshal undelegate callback args")
		}

		// Submit the undelegation ICA
		if _, err := k.SubmitTxsDayEpoch(
			ctx,
			hostZone.ConnectionId,
			msgsBatch,
			types.ICAAccountType_DELEGATION,
			ICACallbackID_Undelegate,
			callbackArgsBz,
		); err != nil {
			return numTxsSubmitted, errorsmod.Wrapf(err, "unable to submit unbonding ICA for %s", hostZone.ChainId)
		}
		numTxsSubmitted += 1

		// flag the delegation change in progress on each validator
		for _, unbonding := range unbondingsBatch {
			if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, unbonding.Validator); err != nil {
				return numTxsSubmitted, err
			}
		}
		k.SetHostZone(ctx, hostZone)
	}

	return numTxsSubmitted, nil
}

// Submits undelegation ICA messages for a given host zone
//
// First, the total unbond amount is determined from the epoch unbonding records
// Then that unbond amount is allowed to cascade across the validators in order of how proportionally
// different their current delegations are from the weight implied target delegation,
// until their capacities have consumed the full amount
// As a result, unbondings lead to a more balanced distribution of stake across validators
//
// Context: Over time, as LSM Liquid stakes are accepted, the total stake managed by the protocol becomes unbalanced
// as liquid stakes are not aligned with the validator weights. This is only rebalanced once per unbonding period
func (k Keeper) UnbondFromHostZone(ctx sdk.Context, hostZone types.HostZone) (err error) {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Preparing MsgUndelegates from the delegation account to each validator"))

	// Confirm the delegation account was registered
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Get the list of relevant records that should unbond
	epochUnbondingRecordIds, epochNumbersToHostZoneUnbondings := k.GetQueuedHostZoneUnbondingRecords(ctx, hostZone.ChainId)

	// Update the native unbond amount on all relevant records
	// The native amount is calculated from the stTokens
	epochNumbersToHostZoneUnbondings, err = k.RefreshUnbondingNativeTokenAmounts(ctx, hostZone.ChainId, epochNumbersToHostZoneUnbondings)
	if err != nil {
		return err
	}

	// Sum the total number of native tokens from the records above that are ready to unbond
	totalNativeUnbondAmount := k.GetTotalUnbondAmount(epochNumbersToHostZoneUnbondings)
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Total unbonded amount: %v%s", totalNativeUnbondAmount, hostZone.HostDenom))

	// If there's nothing to unbond, return and move on to the next host zone
	if totalNativeUnbondAmount.IsZero() {
		return nil
	}

	// Determine the total eligible unbond amount - excluding delegations to validators with a slash query in progress
	totalValidDelegationBeforeUnbonding := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		if !validator.SlashQueryInProgress {
			totalValidDelegationBeforeUnbonding = totalValidDelegationBeforeUnbonding.Add(validator.Delegation)
		}
	}

	// Determine the ideal balanced delegation for each validator after the unbonding
	//   (as if we were to unbond and then rebalance)
	delegationAfterUnbonding := totalValidDelegationBeforeUnbonding.Sub(totalNativeUnbondAmount)
	balancedDelegationsAfterUnbonding, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, delegationAfterUnbonding)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId)
	}

	// Determine the unbond capacity for each validator
	// Each validator can only unbond up to the difference between their current delegation and their balanced delegation
	// The validator's current delegation will be above their balanced delegation if they've received LSM Liquid Stakes
	//   (which is only rebalanced once per unbonding period)
	validatorUnbondCapacity := k.GetValidatorUnbondCapacity(ctx, hostZone.Validators, balancedDelegationsAfterUnbonding)
	if len(validatorUnbondCapacity) == 0 {
		return fmt.Errorf("there are no validators on %s with sufficient unbond capacity", hostZone.ChainId)
	}

	// Sort the unbonding capacity by priority
	// Priority is determined by checking the how proportionally unbalanced each validator is
	// Zero weight validators will come first in the list
	prioritizedUnbondCapacity, err := SortUnbondingCapacityByPriority(validatorUnbondCapacity)
	if err != nil {
		return err
	}

	// Get the undelegation ICA messages and split delegations for the callback
	undelegateBatchSize := int(utils.UintToInt(hostZone.MaxMessagesPerIcaTx))
	msgs, unbondings, err := k.GetUnbondingICAMessages(
		hostZone,
		totalNativeUnbondAmount,
		prioritizedUnbondCapacity,
	)
	if err != nil {
		return err
	}

	// Shouldn't be possible, but if all the validator's had a target unbonding of zero, do not send an ICA
	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Target unbonded amount was 0 for each validator")
	}

	// Send the messages in batches so the gas limit isn't exceedeed
	numTxsSubmitted, err := k.BatchSubmitUndelegateICAMessages(
		ctx,
		hostZone,
		epochUnbondingRecordIds,
		msgs,
		unbondings,
		undelegateBatchSize,
	)
	if err != nil {
		return err
	}

	// Update the epoch unbonding record status and number of undelegation ICAs
	for _, epochNumber := range utils.Uint64MapKeys(epochNumbersToHostZoneUnbondings) {
		hostZoneUnbonding := epochNumbersToHostZoneUnbondings[epochNumber]
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS
		hostZoneUnbonding.UndelegationTxsInProgress += numTxsSubmitted
		err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, hostZone.ChainId, hostZoneUnbonding)
		if err != nil {
			return err
		}
	}

	EmitUndelegationEvent(ctx, hostZone, totalNativeUnbondAmount)

	return nil
}

// this function iterates each host zone, and if it's the right time to
// initiate an unbonding, it attempts to unbond all outstanding records
func (k Keeper) InitiateAllHostZoneUnbondings(ctx sdk.Context, dayNumber uint64) {
	k.Logger(ctx).Info(fmt.Sprintf("Initiating all host zone unbondings for epoch %d...", dayNumber))

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {

		// Confirm the unbonding is supposed to be triggered this epoch
		unbondingFrequency := hostZone.GetUnbondingFrequency()
		if dayNumber%unbondingFrequency != 0 {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
				"Host does not unbond this epoch (Unbonding Period: %d, Unbonding Frequency: %d, Epoch: %d)",
				hostZone.UnbondingPeriod, unbondingFrequency, dayNumber))
			continue
		}

		// Get host zone unbonding message by summing up the unbonding records
		err := utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			return k.UnbondFromHostZone(ctx, hostZone)
		})
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error initiating host zone unbondings for host zone %s: %s", hostZone.ChainId, err.Error()))
			continue
		}
	}
}
