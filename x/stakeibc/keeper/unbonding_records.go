package keeper

import (
	"errors"
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v14/utils"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const (
	UndelegateICABatchSize = 30
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

// Creates a new epoch unbonding record for the epoch
func (k Keeper) CreateEpochUnbondingRecord(ctx sdk.Context, epochNumber uint64) bool {
	k.Logger(ctx).Info(fmt.Sprintf("Creating Epoch Unbonding Records for Epoch %d", epochNumber))

	hostZoneUnbondings := []*recordstypes.HostZoneUnbonding{}

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Creating Epoch Unbonding Record"))

		hostZoneUnbonding := recordstypes.HostZoneUnbonding{
			NativeTokenAmount: sdkmath.ZeroInt(),
			StTokenAmount:     sdkmath.ZeroInt(),
			Denom:             hostZone.HostDenom,
			HostZoneId:        hostZone.ChainId,
			Status:            recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,
		}
		hostZoneUnbondings = append(hostZoneUnbondings, &hostZoneUnbonding)
	}

	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		EpochNumber:        cast.ToUint64(epochNumber),
		HostZoneUnbondings: hostZoneUnbondings,
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return true
}

// Gets the total unbonded amount for a host zone by looping through the epoch unbonding records
// Also returns the epoch unbonding record ids
func (k Keeper) GetTotalUnbondAmountAndRecordsIds(ctx sdk.Context, chainId string) (totalUnbonded sdkmath.Int, unbondingRecordIds []uint64) {
	totalUnbonded = sdk.ZeroInt()
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbonding.EpochNumber, chainId)
		if !found {
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Epoch %d - Status: %s, Amount: %v",
			epochUnbonding.EpochNumber, hostZoneRecord.Status, hostZoneRecord.NativeTokenAmount))

		// We'll unbond all records that have status UNBONDING_QUEUE and have an amount g.t. zero
		if hostZoneRecord.Status == recordstypes.HostZoneUnbonding_UNBONDING_QUEUE && hostZoneRecord.NativeTokenAmount.GT(sdkmath.ZeroInt()) {
			totalUnbonded = totalUnbonded.Add(hostZoneRecord.NativeTokenAmount)
			unbondingRecordIds = append(unbondingRecordIds, epochUnbonding.EpochNumber)
			k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  %v%s included in total unbonding", hostZoneRecord.NativeTokenAmount, hostZoneRecord.Denom))
		}
	}
	return totalUnbonded, unbondingRecordIds
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
) (msgs []proto.Message, unbondings []*types.SplitDelegation, err error) {
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
		unbondToken := sdk.NewCoin(hostZone.HostDenom, unbondAmount)

		// Build the undelegate ICA messages
		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validatorCapacity.ValidatorAddress,
			Amount:           unbondToken,
		})

		// Build the validator splits for the callback
		unbondings = append(unbondings, &types.SplitDelegation{
			Validator: validatorCapacity.ValidatorAddress,
			Amount:    unbondAmount,
		})
	}

	// Sanity check that we had enough capacity to unbond
	if !remainingUnbondAmount.IsZero() {
		return msgs, unbondings,
			fmt.Errorf("unable to unbond full amount (%v) from %v", totalUnbondAmount, hostZone.ChainId)
	}

	return msgs, unbondings, nil
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
func (k Keeper) UnbondFromHostZone(ctx sdk.Context, hostZone types.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Preparing MsgUndelegates from the delegation account to each validator"))

	// Confirm the delegation account was registered
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Iterate through every unbonding record and sum the total amount to unbond for the given host zone
	totalUnbondAmount, epochUnbondingRecordIds := k.GetTotalUnbondAmountAndRecordsIds(ctx, hostZone.ChainId)
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Total unbonded amount: %v%s", totalUnbondAmount, hostZone.HostDenom))

	// If there's nothing to unbond, return and move on to the next host zone
	if totalUnbondAmount.IsZero() {
		return nil
	}

	// Determine the ideal balanced delegation for each validator after the unbonding
	//   (as if we were to unbond and then rebalance)
	// This will serve as the starting point for determining how much to unbond each validator
	delegationAfterUnbonding := hostZone.TotalDelegations.Sub(totalUnbondAmount)
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
	msgs, unbondings, err := k.GetUnbondingICAMessages(hostZone, totalUnbondAmount, prioritizedUnbondCapacity)
	if err != nil {
		return err
	}

	// Shouldn't be possible, but if all the validator's had a target unbonding of zero, do not send an ICA
	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Target unbonded amount was 0 for each validator")
	}

	// Send the messages in batches so the gas limit isn't exceedeed
	for start := 0; start < len(msgs); start += UndelegateICABatchSize {
		end := start + UndelegateICABatchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		msgsBatch := msgs[start:end]
		unbondingsBatch := unbondings[start:end]

		// Store the callback data
		undelegateCallback := types.UndelegateCallback{
			HostZoneId:              hostZone.ChainId,
			SplitDelegations:        unbondingsBatch,
			EpochUnbondingRecordIds: epochUnbondingRecordIds,
		}
		callbackArgsBz, err := proto.Marshal(&undelegateCallback)
		if err != nil {
			return errorsmod.Wrap(err, "unable to marshal undelegate callback args")
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
			return errorsmod.Wrapf(err, "unable to submit unbonding ICA for %s", hostZone.ChainId)
		}

		// flag the delegation change in progress on each validator
		for _, unbonding := range unbondingsBatch {
			if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, unbonding.Validator); err != nil {
				return err
			}
		}
		k.SetHostZone(ctx, hostZone)
	}

	// Update the epoch unbonding record status
	if err := k.RecordsKeeper.SetHostZoneUnbondings(
		ctx,
		hostZone.ChainId,
		epochUnbondingRecordIds,
		recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
	); err != nil {
		return err
	}

	EmitUndelegationEvent(ctx, hostZone, totalUnbondAmount)

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
		if err := k.UnbondFromHostZone(ctx, hostZone); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error initiating host zone unbondings for host zone %s: %s", hostZone.ChainId, err.Error()))
			continue
		}
	}
}

// Deletes any epoch unbonding records that have had all unbondings claimed
func (k Keeper) CleanupEpochUnbondingRecords(ctx sdk.Context, epochNumber uint64) bool {
	k.Logger(ctx).Info("Cleaning Claimed Epoch Unbonding Records...")

	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		shouldDeleteEpochUnbondingRecord := true
		hostZoneUnbondings := epochUnbondingRecord.HostZoneUnbondings

		for _, hostZoneUnbonding := range hostZoneUnbondings {
			// if an EpochUnbondingRecord has any HostZoneUnbonding with non-zero balances, we don't delete the EpochUnbondingRecord
			// because it has outstanding tokens that need to be claimed
			if !hostZoneUnbonding.NativeTokenAmount.Equal(sdkmath.ZeroInt()) {
				shouldDeleteEpochUnbondingRecord = false
				break
			}
		}
		if shouldDeleteEpochUnbondingRecord {
			k.Logger(ctx).Info(fmt.Sprintf("  EpochUnbondingRecord %d - All unbondings claimed, removing record", epochUnbondingRecord.EpochNumber))
			k.RecordsKeeper.RemoveEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber)
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("  EpochUnbondingRecord %d - Has unclaimed unbondings", epochUnbondingRecord.EpochNumber))
		}
	}

	return true
}

// Batch transfers any unbonded tokens from the delegation account to the redemption account
func (k Keeper) SweepAllUnbondedTokensForHostZone(ctx sdk.Context, hostZone types.HostZone, epochUnbondingRecords []recordstypes.EpochUnbondingRecord) (success bool, sweepAmount sdkmath.Int) {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Sweeping unbonded tokens"))

	// Sum up all host zone unbonding records that have finished unbonding
	totalAmtTransferToRedemptionAcct := sdkmath.ZeroInt()
	epochUnbondingRecordIds := []uint64{}
	for _, epochUnbondingRecord := range epochUnbondingRecords {

		// Get all the unbondings associated with the epoch + host zone pair
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
		if !found {
			continue
		}

		// Get latest blockTime from light client
		blockTime, err := k.GetLightClientTimeSafely(ctx, hostZone.ConnectionId)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("\tCould not find blockTime for host zone %s", hostZone.ChainId))
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Epoch %d - Status: %s, Amount: %v, Unbonding Time: %d, Block Time: %d",
			epochUnbondingRecord.EpochNumber, hostZoneUnbonding.Status.String(), hostZoneUnbonding.NativeTokenAmount, hostZoneUnbonding.UnbondingTime, blockTime))

		// If the unbonding period has elapsed, then we can send the ICA call to sweep this
		//   hostZone's unbondings to the redemption account (in a batch).
		// Verify:
		//      1. the unbonding time is set (g.t. 0)
		//      2. the unbonding time is less than the current block time
		//      3. the host zone is in the EXIT_TRANSFER_QUEUE state, meaning it's ready to be transferred
		inTransferQueue := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
		validUnbondingTime := hostZoneUnbonding.UnbondingTime > 0 && hostZoneUnbonding.UnbondingTime < blockTime
		if inTransferQueue && validUnbondingTime {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "  %v%s included in sweep", hostZoneUnbonding.NativeTokenAmount, hostZoneUnbonding.Denom))

			if err != nil {
				errMsg := fmt.Sprintf("Could not convert native token amount to int64 | %s", err.Error())
				k.Logger(ctx).Error(errMsg)
				continue
			}
			totalAmtTransferToRedemptionAcct = totalAmtTransferToRedemptionAcct.Add(hostZoneUnbonding.NativeTokenAmount)
			epochUnbondingRecordIds = append(epochUnbondingRecordIds, epochUnbondingRecord.EpochNumber)
		}
	}

	// If we have any amount to sweep, then we can send the ICA call to sweep them
	if totalAmtTransferToRedemptionAcct.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "No tokens ready for sweep"))
		return true, totalAmtTransferToRedemptionAcct
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Batch transferring %v to host zone", totalAmtTransferToRedemptionAcct))

	// Get the delegation account and redemption account
	if hostZone.DelegationIcaAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return false, sdkmath.ZeroInt()
	}
	if hostZone.RedemptionIcaAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a redemption address!", hostZone.ChainId))
		return false, sdkmath.ZeroInt()
	}

	// Build transfer message to transfer from the delegation account to redemption account
	sweepCoin := sdk.NewCoin(hostZone.HostDenom, totalAmtTransferToRedemptionAcct)
	msgs := []proto.Message{
		&banktypes.MsgSend{
			FromAddress: hostZone.DelegationIcaAddress,
			ToAddress:   hostZone.RedemptionIcaAddress,
			Amount:      sdk.NewCoins(sweepCoin),
		},
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Preparing MsgSend from Delegation Account to Redemption Account"))

	// Store the epoch numbers in the callback to identify the epoch unbonding records
	redemptionCallback := types.RedemptionCallback{
		HostZoneId:              hostZone.ChainId,
		EpochUnbondingRecordIds: epochUnbondingRecordIds,
	}
	marshalledCallbackArgs, err := k.MarshalRedemptionCallbackArgs(ctx, redemptionCallback)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return false, sdkmath.ZeroInt()
	}

	// Send the transfer ICA
	_, err = k.SubmitTxsDayEpoch(ctx, hostZone.ConnectionId, msgs, types.ICAAccountType_DELEGATION, ICACallbackID_Redemption, marshalledCallbackArgs)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to SubmitTxs, transfer to redemption account on %s", hostZone.ChainId))
		return false, sdkmath.ZeroInt()
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgSend Successfully Sent"))

	// Update the host zone unbonding records to status IN_PROGRESS
	err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, hostZone.ChainId, epochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return false, sdkmath.ZeroInt()
	}

	return true, totalAmtTransferToRedemptionAcct
}

// Sends all unbonded tokens to the redemption account
// returns:
//   - success indicator if all chains succeeded
//   - list of successful chains
//   - list of tokens swept
//   - list of failed chains
func (k Keeper) SweepAllUnbondedTokens(ctx sdk.Context) (success bool, successfulSweeps []string, sweepAmounts []sdkmath.Int, failedSweeps []string) {
	// this function returns true if all chains succeeded, false otherwise
	// it also returns a list of successful chains (arg 2), tokens swept (arg 3), and failed chains (arg 4)
	k.Logger(ctx).Info("Sweeping All Unbonded Tokens...")

	success = true
	successfulSweeps = []string{}
	sweepAmounts = []sdkmath.Int{}
	failedSweeps = []string{}
	hostZones := k.GetAllActiveHostZone(ctx)

	epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
	for _, hostZone := range hostZones {
		hostZoneSuccess, sweepAmount := k.SweepAllUnbondedTokensForHostZone(ctx, hostZone, epochUnbondingRecords)
		if hostZoneSuccess {
			successfulSweeps = append(successfulSweeps, hostZone.ChainId)
			sweepAmounts = append(sweepAmounts, sweepAmount)
		} else {
			success = false
			failedSweeps = append(failedSweeps, hostZone.ChainId)
		}
	}

	return success, successfulSweeps, sweepAmounts, failedSweeps
}
