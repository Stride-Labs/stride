package keeper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
	vestingtypes "github.com/Stride-Labs/stride/v9/x/claim/vesting/types"
	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
)

func (k Keeper) LoadAllocationData(ctx sdk.Context, allocationData string) bool {
	records := []types.ClaimRecord{}

	lines := strings.Split(allocationData, "\n")
	allocatedFlags := map[string]bool{}
	for _, line := range lines {
		data := strings.Split(line, ",")
		if data[0] == "" || data[1] == "" || data[2] == "" {
			continue
		}

		airdropIdentifier := data[0]
		sourceChainAddr := data[1]
		airdropWeight := data[2]
		strideAddr := utils.ConvertAddressToStrideAddress(sourceChainAddr)
		if strideAddr == "" {
			continue
		}
		allocationIdentifier := airdropIdentifier + strideAddr

		// Round weight value so that it always has 10 decimal places
		weightFloat64, err := strconv.ParseFloat(airdropWeight, 64)
		if err != nil {
			continue
		}

		weightStr := fmt.Sprintf("%.10f", weightFloat64)
		weight, err := sdk.NewDecFromStr(weightStr)
		if weight.IsNegative() || weight.IsZero() {
			continue
		}

		if err != nil || allocatedFlags[allocationIdentifier] {
			continue
		}

		_, err = sdk.AccAddressFromBech32(strideAddr)
		if err != nil {
			continue
		}

		records = append(records, types.ClaimRecord{
			AirdropIdentifier: airdropIdentifier,
			Address:           strideAddr,
			Weight:            weight,
			ActionCompleted:   []bool{false, false, false},
		})

		allocatedFlags[allocationIdentifier] = true
	}

	if err := k.SetClaimRecordsWithWeights(ctx, records); err != nil {
		panic(err)
	}
	return true
}

// Remove duplicated airdrops for given params
func (k Keeper) GetUnallocatedUsers(ctx sdk.Context, identifier string, users []string, weights []sdk.Dec) ([]string, []sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(identifier)...))
	newUsers := []string{}
	newWeights := []sdk.Dec{}
	for idx, user := range users {
		strideAddr := utils.ConvertAddressToStrideAddress(user)
		addr, _ := sdk.AccAddressFromBech32(strideAddr)
		// If new user, then append user and weight
		if !prefixStore.Has(addr) {
			newUsers = append(newUsers, user)
			newWeights = append(newWeights, weights[idx])
		}
	}

	return newUsers, newWeights
}

// Get airdrop duration for action
func GetAirdropDurationForAction(action types.Action) int64 {
	if action == types.ACTION_DELEGATE_STAKE {
		return int64(types.DefaultVestingDurationForDelegateStake.Seconds())
	} else if action == types.ACTION_LIQUID_STAKE {
		return int64(types.DefaultVestingDurationForLiquidStake.Seconds())
	}
	return int64(0)
}

// Get airdrop by distributor
func (k Keeper) GetAirdropByDistributor(ctx sdk.Context, distributor string) *types.Airdrop {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	if distributor == "" {
		return nil
	}

	for _, airdrop := range params.Airdrops {
		if airdrop.DistributorAddress == distributor {
			return airdrop
		}
	}

	return nil
}

// Get airdrop by identifier
func (k Keeper) GetAirdropByIdentifier(ctx sdk.Context, airdropIdentifier string) *types.Airdrop {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	if airdropIdentifier == "" {
		return nil
	}

	for _, airdrop := range params.Airdrops {
		if airdrop.AirdropIdentifier == airdropIdentifier {
			return airdrop
		}
	}

	return nil
}

// Get airdrop by chain id
func (k Keeper) GetAirdropByChainId(ctx sdk.Context, chainId string) (airdrop *types.Airdrop, found bool) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	if chainId == "" {
		return nil, false
	}

	for _, airdrop := range params.Airdrops {
		if airdrop.ChainId == chainId {
			return airdrop, true
		}
	}

	return nil, false
}

func (k Keeper) GetAirdropIds(ctx sdk.Context) []string {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// init airdrop ids
	airdropIds := []string{}

	for _, airdrop := range params.Airdrops {
		// append airdrop to airdrop ids
		airdropIds = append(airdropIds, airdrop.AirdropIdentifier)
	}

	return airdropIds
}

// GetDistributorAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetDistributorAccountBalance(ctx sdk.Context, airdropIdentifier string) (sdk.Coin, error) {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid airdrop identifier: GetDistributorAccountBalance")
	}

	addr, err := k.GetAirdropDistributor(ctx, airdropIdentifier)
	if err != nil {
		return sdk.Coin{}, err
	}
	return k.bankKeeper.GetBalance(ctx, addr, airdrop.ClaimDenom), nil
}

// EndAirdrop ends airdrop and clear all user claim records
func (k Keeper) EndAirdrop(ctx sdk.Context, airdropIdentifier string) error {
	ctx.Logger().Info("Clearing claims module state entries")
	k.clearInitialClaimables(ctx, airdropIdentifier)
	k.DeleteTotalWeight(ctx, airdropIdentifier)
	return k.DeleteAirdropAndEpoch(ctx, airdropIdentifier)
}

func (k Keeper) IsInitialPeriodPassed(ctx sdk.Context, airdropIdentifier string) bool {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		k.Logger(ctx).Info("[CLAIM] airdrop is nil")
		return false
	}
	goneTime := ctx.BlockTime().Sub(airdrop.AirdropStartTime)
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] goneTime %v", goneTime))
	// Check if elapsed time since airdrop start is over the initial period of vesting
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] goneTime.Seconds() %v", goneTime.Seconds()))
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] types.DefaultVestingInitialPeriod.Seconds() %v", types.DefaultVestingInitialPeriod.Seconds()))
	return goneTime.Seconds() >= types.DefaultVestingInitialPeriod.Seconds()
}

// ResetClaimStatus clear users' claimed status only after initial period of vesting is passed
func (k Keeper) ResetClaimStatus(ctx sdk.Context, airdropIdentifier string) error {
	passed := k.IsInitialPeriodPassed(ctx, airdropIdentifier)
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] k.IsInitialPeriodPassed(ctx, airdropIdentifier) %v", passed))
	if passed {
		k.Logger(ctx).Info("Resetting claim status")
		// first, reset the claim records
		records := k.GetClaimRecords(ctx, airdropIdentifier)
		k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] len(records) %v", len(records)))
		for idx := range records {
			records[idx].ActionCompleted = []bool{false, false, false}
		}

		k.Logger(ctx).Info("[CLAIM] SetClaimRecords...")
		if err := k.SetClaimRecords(ctx, records); err != nil {
			k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] SetClaimRecords  %v", err.Error()))
			return err
		}
		// then, reset the airdrop ClaimedSoFar
		k.Logger(ctx).Info("[CLAIM] ResetClaimedSoFar...")
		if err := k.ResetClaimedSoFar(ctx, airdropIdentifier); err != nil {
			k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] ResetClaimedSoFar %v", err.Error()))
			return err
		}
	}
	return nil
}

// ClearClaimables clear claimable amounts
func (k Keeper) clearInitialClaimables(ctx sdk.Context, airdropIdentifier string) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(airdropIdentifier)...))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}
}

func (k Keeper) SetClaimRecordsWithWeights(ctx sdk.Context, claimRecords []types.ClaimRecord) error {
	// Set total weights
	weights := make(map[string]sdk.Dec)
	for _, record := range claimRecords {
		if weights[record.AirdropIdentifier].IsNil() {
			weights[record.AirdropIdentifier] = sdk.ZeroDec()
		}

		weights[record.AirdropIdentifier] = weights[record.AirdropIdentifier].Add(record.Weight)
	}

	// DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
	for _, identifier := range utils.StringMapKeys(weights) {
		weight := weights[identifier]
		k.SetTotalWeight(ctx, weight, identifier)
	}

	// Set claim records
	return k.SetClaimRecords(ctx, claimRecords)
}

// SetClaimRecords set claim records and total weights
func (k Keeper) SetClaimRecords(ctx sdk.Context, claimRecords []types.ClaimRecord) error {
	for _, claimRecord := range claimRecords {
		err := k.SetClaimRecord(ctx, claimRecord)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetClaimables get claimables for genesis export
func (k Keeper) GetClaimRecords(ctx sdk.Context, airdropIdentifier string) []types.ClaimRecord {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(airdropIdentifier)...))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	claimRecords := []types.ClaimRecord{}
	for ; iterator.Valid(); iterator.Next() {

		claimRecord := types.ClaimRecord{}

		err := proto.Unmarshal(iterator.Value(), &claimRecord)
		if err != nil {
			panic(err)
		}

		claimRecords = append(claimRecords, claimRecord)
	}
	return claimRecords
}

// GetClaimRecord returns the claim record for a specific address
func (k Keeper) GetClaimRecord(ctx sdk.Context, addr sdk.AccAddress, airdropIdentifier string) (types.ClaimRecord, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(airdropIdentifier)...))
	if !prefixStore.Has(addr) {
		return types.ClaimRecord{}, nil
	}
	bz := prefixStore.Get(addr)

	claimRecord := types.ClaimRecord{}
	err := proto.Unmarshal(bz, &claimRecord)
	if err != nil {
		return types.ClaimRecord{}, err
	}

	return claimRecord, nil
}

// SetTotalWeight sets total sum of user weights in store
func (k Keeper) SetTotalWeight(ctx sdk.Context, totalWeight sdk.Dec, airdropIdentifier string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(append([]byte(types.TotalWeightKey), []byte(airdropIdentifier)...), []byte(totalWeight.String()))
}

// DeleteTotalWeight deletes total weights for airdrop
func (k Keeper) DeleteTotalWeight(ctx sdk.Context, airdropIdentifier string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(append([]byte(types.TotalWeightKey), []byte(airdropIdentifier)...))
}

// GetTotalWeight gets total sum of user weights in store
func (k Keeper) GetTotalWeight(ctx sdk.Context, airdropIdentifier string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(append([]byte(types.TotalWeightKey), []byte(airdropIdentifier)...))
	if b == nil {
		return sdk.ZeroDec(), nil
	}
	totalWeight, err := sdk.NewDecFromStr(string(b))
	if err != nil {
		return sdk.ZeroDec(), types.ErrTotalWeightParse
	}
	return totalWeight, nil
}

// SetClaimRecord sets a claim record for an address in store
func (k Keeper) SetClaimRecord(ctx sdk.Context, claimRecord types.ClaimRecord) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(claimRecord.AirdropIdentifier)...))

	bz, err := proto.Marshal(&claimRecord)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(claimRecord.Address)
	if err != nil {
		return err
	}

	prefixStore.Set(addr, bz)
	return nil
}

func (k Keeper) DeleteClaimRecord(ctx sdk.Context, addr sdk.AccAddress, airdropId string) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(airdropId)...))

	prefixStore.Delete(addr)

	return nil
}

// Get airdrop distributor address
func (k Keeper) GetAirdropDistributor(ctx sdk.Context, airdropIdentifier string) (sdk.AccAddress, error) {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return sdk.AccAddress{}, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid airdrop identifier: GetAirdropDistributor")
	}
	return sdk.AccAddressFromBech32(airdrop.DistributorAddress)
}

// Get airdrop claim denom
func (k Keeper) GetAirdropClaimDenom(ctx sdk.Context, airdropIdentifier string) (string, error) {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid airdrop identifier: GetAirdropClaimDenom")
	}
	return airdrop.ClaimDenom, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action, airdropIdentifier string, includeClaimed bool) (sdk.Coins, error) {
	claimRecord, err := k.GetClaimRecord(ctx, addr, airdropIdentifier)
	if err != nil {
		return nil, err
	}

	if claimRecord.Address == "" {
		return sdk.Coins{}, nil
	}

	// if action already completed (and we're not including claimed tokens), nothing is claimable
	if !includeClaimed && claimRecord.ActionCompleted[action] {
		return sdk.Coins{}, nil
	}

	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if ctx.BlockTime().Before(airdrop.AirdropStartTime) {
		return sdk.Coins{}, nil
	}

	totalWeight, err := k.GetTotalWeight(ctx, airdropIdentifier)
	if err != nil {
		return nil, types.ErrFailedToGetTotalWeight
	}

	percentageForAction := types.PercentageForFree
	if action == types.ACTION_DELEGATE_STAKE {
		percentageForAction = types.PercentageForStake
	} else if action == types.ACTION_LIQUID_STAKE {
		percentageForAction = types.PercentageForLiquidStake
	}

	distributorAccountBalance, err := k.GetDistributorAccountBalance(ctx, airdropIdentifier)
	if err != nil {
		return sdk.Coins{}, err
	}

	poolBal := distributorAccountBalance.AddAmount(airdrop.ClaimedSoFar)

	claimableAmount := sdk.NewDec(poolBal.Amount.Int64()).
		Mul(percentageForAction).
		Mul(claimRecord.Weight).
		Quo(totalWeight).RoundInt()
	claimableCoins := sdk.NewCoins(sdk.NewCoin(airdrop.ClaimDenom, claimableAmount))

	elapsedAirdropTime := ctx.BlockTime().Sub(airdrop.AirdropStartTime)
	// The entire airdrop has completed
	if elapsedAirdropTime > airdrop.AirdropDuration {
		return sdk.Coins{}, nil
	}
	return claimableCoins, nil
}

// GetUserVestings returns all vestings associated to the user account
func (k Keeper) GetUserVestings(ctx sdk.Context, addr sdk.AccAddress) (vestingtypes.Periods, sdk.Coins) {
	acc := k.accountKeeper.GetAccount(ctx, addr)
	strideVestingAcc, isStrideVestingAccount := acc.(*vestingtypes.StridePeriodicVestingAccount)
	if !isStrideVestingAccount {
		return vestingtypes.Periods{}, sdk.Coins{}
	} else {
		return strideVestingAcc.VestingPeriods, strideVestingAcc.GetVestedCoins(ctx.BlockTime())
	}
}

func AreAllTrue(bools []bool) bool {
	for _, b := range bools {
		if !b {
			return false
		}
	}
	return true
}

// GetClaimStatus returns all claim status associated with the user account
func (k Keeper) GetClaimStatus(ctx sdk.Context, addr sdk.AccAddress) ([]types.ClaimStatus, error) {
	// Get all airdrop identifiers
	airdropIdentifiers := k.GetAirdropIds(ctx)
	var claimStatusList []types.ClaimStatus
	for _, airdropId := range airdropIdentifiers {

		// Get the claim record for a user, airdrop pair
		claimRecord, err := k.GetClaimRecord(ctx, addr, airdropId)
		if err != nil {
			return nil, err
		}
		if claimRecord.Address == "" {
			// if there's no claim record, the user is not eligible
			// for this airdrop, so skip it
			continue
		}

		// If all actions are completed, the user has claimed
		claimed := AreAllTrue(claimRecord.ActionCompleted)
		claimStatus := types.ClaimStatus{
			AirdropIdentifier: airdropId,
			Claimed:           claimed,
		}
		claimStatusList = append(claimStatusList, claimStatus)
	}

	return claimStatusList, nil
}

func CurrentAirdropRound(start time.Time) int {
	// Define constants for 90 days and 30 days
	const initialRoundDuration = 90 * 24 * time.Hour
	const subsequentRoundDuration = 30 * 24 * time.Hour

	// Calculate the time passed since the start
	timePassed := time.Since(start)

	// Check if the initial round is still ongoing
	if timePassed < initialRoundDuration {
		return 1
	}

	// Calculate the time passed after the initial round
	timePassedAfterInitialRound := timePassed - initialRoundDuration

	// Calculate the number of subsequent rounds passed
	subsequentRoundsPassed := timePassedAfterInitialRound / subsequentRoundDuration

	// Add 1 for the initial round and 1 for the current round
	return 1 + 1 + int(subsequentRoundsPassed)
}

// GetClaimMetadata returns all claim status associated with the user account
func (k Keeper) GetClaimMetadata(ctx sdk.Context) []types.ClaimMetadata {
	var claimMetadataList []types.ClaimMetadata

	airdropIdentifiers := k.GetAirdropIds(ctx)
	epochs := k.epochsKeeper.AllEpochInfos(ctx)

	for _, airdropId := range airdropIdentifiers {
		// loop over epochs to match epochs to airdrop identifier
		var currentRoundStart time.Time
		var currentRoundEnd time.Time
		var absoluteStartTime time.Time
		var duration time.Duration
		for _, epoch := range epochs {
			epochIdentifier := strings.TrimPrefix(epoch.Identifier, "airdrop-")
			if epochIdentifier == airdropId {
				// found the epoch for this airdrop
				currentRoundStart = epoch.CurrentEpochStartTime
				absoluteStartTime = epoch.StartTime
				duration = epoch.Duration
			}
		}

		currentRoundEnd = currentRoundStart.Add(duration)
		currentRound := strconv.Itoa(CurrentAirdropRound(absoluteStartTime))

		claimMetadata := types.ClaimMetadata{
			AirdropIdentifier: airdropId,
			CurrentRound:      currentRound,
			CurrentRoundStart: currentRoundStart,
			CurrentRoundEnd:   currentRoundEnd,
		}

		claimMetadataList = append(claimMetadataList, claimMetadata)
	}

	return claimMetadataList
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress, airdropIdentifier string, includeClaimed bool) (sdk.Coins, error) {
	claimRecord, err := k.GetClaimRecord(ctx, addr, airdropIdentifier)
	if err != nil {
		return sdk.Coins{}, err
	}
	if claimRecord.Address == "" {
		return sdk.Coins{}, nil
	}

	totalClaimable := sdk.Coins{}

	for action := range utils.Int32MapKeys(types.Action_name) {
		claimableForAction, err := k.GetClaimableAmountForAction(ctx, addr, types.Action(action), airdropIdentifier, includeClaimed)
		if err != nil {
			return sdk.Coins{}, err
		}
		if !claimableForAction.Empty() {
			totalClaimable = totalClaimable.Add(claimableForAction...)
		}
	}
	return totalClaimable, nil
}

// Get airdrop identifier corresponding to the user address
func (k Keeper) GetAirdropIdentifiersForUser(ctx sdk.Context, addr sdk.AccAddress) []string {
	store := ctx.KVStore(k.storeKey)
	params, err := k.GetParams(ctx)
	identifiers := []string{}
	if err != nil {
		return identifiers
	}

	for _, airdrop := range params.Airdrops {
		prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(airdrop.AirdropIdentifier)...))
		if prefixStore.Has(addr) {
			identifiers = append(identifiers, airdrop.AirdropIdentifier)
		}
	}
	return identifiers
}

func (k Keeper) AfterClaim(ctx sdk.Context, airdropIdentifier string, claimAmount sdkmath.Int) error {
	// Increment ClaimedSoFar on the airdrop record
	// fetch the airdrop
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid airdrop identifier: AfterClaim")
	}
	// increment the claimed so far
	err := k.IncrementClaimedSoFar(ctx, airdropIdentifier, claimAmount)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) ClaimAllCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error) {
	// get all airdrops for the user
	airdropIdentifiers := k.GetAirdropIdentifiersForUser(ctx, addr)
	// claim all coins for the action
	totalClaimable := sdk.Coins{}
	for _, airdropIdentifier := range airdropIdentifiers {
		claimable, err := k.ClaimCoinsForAction(ctx, addr, action, airdropIdentifier)
		if err != nil {
			return sdk.Coins{}, err
		}
		totalClaimable = totalClaimable.Add(claimable...)
	}
	return totalClaimable, nil
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action, airdropIdentifier string) (sdk.Coins, error) {
	isPassed := k.IsInitialPeriodPassed(ctx, airdropIdentifier)
	if airdropIdentifier == "" {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid airdrop identifier: ClaimCoinsForAction")
	}

	claimableAmount, err := k.GetClaimableAmountForAction(ctx, addr, action, airdropIdentifier, false)
	if err != nil {
		return claimableAmount, err
	}

	if claimableAmount.Empty() {
		return claimableAmount, nil
	}

	claimRecord, err := k.GetClaimRecord(ctx, addr, airdropIdentifier)
	if err != nil {
		return nil, err
	}

	// Only BaseAccounts and StridePeriodicVestingAccount can claim
	acc := k.accountKeeper.GetAccount(ctx, addr)
	_, isStrideVestingAccount := acc.(*vestingtypes.StridePeriodicVestingAccount)
	_, isBaseAcc := acc.(*authtypes.BaseAccount)
	canClaim := isStrideVestingAccount || isBaseAcc
	if !canClaim {
		return nil, errorsmod.Wrapf(types.ErrInvalidAccount, "Account: %v", acc)
	}

	// Claims don't vest if action type is ActionFree or initial period of vesting is passed
	if !isPassed {
		acc = k.accountKeeper.GetAccount(ctx, addr)
		strideVestingAcc, isStrideVestingAccount := acc.(*vestingtypes.StridePeriodicVestingAccount)
		// Check if vesting tokens already exist for this account.
		if !isStrideVestingAccount {
			// Convert user account into stride veting account.
			baseAccount := k.accountKeeper.NewAccountWithAddress(ctx, addr)
			if _, ok := baseAccount.(*authtypes.BaseAccount); !ok {
				return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid account type; expected: BaseAccount, got: %T", baseAccount)
			}

			periodLength := GetAirdropDurationForAction(action)
			vestingAcc := vestingtypes.NewStridePeriodicVestingAccount(baseAccount.(*authtypes.BaseAccount), claimableAmount, []vestingtypes.Period{{
				StartTime:  ctx.BlockTime().Unix(),
				Length:     periodLength,
				Amount:     claimableAmount,
				ActionType: int32(action),
			}})
			k.accountKeeper.SetAccount(ctx, vestingAcc)
		} else {
			// Grant a new vesting to the existing stride vesting account
			periodLength := GetAirdropDurationForAction(action)
			strideVestingAcc.AddNewGrant(vestingtypes.Period{
				StartTime:  ctx.BlockTime().Unix(),
				Length:     periodLength,
				Amount:     claimableAmount,
				ActionType: int32(action),
			})
			k.accountKeeper.SetAccount(ctx, strideVestingAcc)
		}
	}

	distributor, err := k.GetAirdropDistributor(ctx, airdropIdentifier)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoins(ctx, distributor, addr, claimableAmount)
	if err != nil {
		return nil, err
	}

	claimRecord.ActionCompleted[action] = true

	err = k.SetClaimRecord(ctx, claimRecord)
	if err != nil {
		return claimableAmount, err
	}

	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid airdrop identifier: ClaimCoinsForAction")
	}
	err = k.AfterClaim(ctx, airdropIdentifier, claimableAmount.AmountOf(airdrop.ClaimDenom))
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimableAmount.String()),
		),
	})

	return claimableAmount, nil
}

// CreateAirdropAndEpoch creates a new airdrop and epoch for that.
func (k Keeper) CreateAirdropAndEpoch(ctx sdk.Context, msg types.MsgCreateAirdrop) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// re-run validate basic in case this function is called directly from an upgrade handler
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	for _, airdrop := range params.Airdrops {
		if airdrop.AirdropIdentifier == msg.Identifier {
			return types.ErrAirdropAlreadyExists
		}
		if airdrop.ChainId == msg.ChainId {
			return types.ErrAirdropChainIdAlreadyExists
		}
	}

	airdrop := types.Airdrop{
		AirdropIdentifier:  msg.Identifier,
		ChainId:            msg.ChainId,
		AirdropDuration:    time.Duration(msg.Duration * uint64(time.Second)),
		ClaimDenom:         msg.Denom,
		DistributorAddress: msg.Distributor,
		AirdropStartTime:   time.Unix(int64(msg.StartTime), 0),
		AutopilotEnabled:   msg.AutopilotEnabled,
	}

	params.Airdrops = append(params.Airdrops, &airdrop)
	k.epochsKeeper.SetEpochInfo(ctx, epochstypes.EpochInfo{
		Identifier:              fmt.Sprintf("airdrop-%s", msg.Identifier),
		StartTime:               airdrop.AirdropStartTime.Add(time.Minute),
		Duration:                types.DefaultEpochDuration,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	})
	return k.SetParams(ctx, params)
}

// IncrementClaimedSoFar increments ClaimedSoFar for a single airdrop
func (k Keeper) IncrementClaimedSoFar(ctx sdk.Context, identifier string, amount sdkmath.Int) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	if amount.LT(sdkmath.ZeroInt()) {
		return types.ErrInvalidAmount
	}

	newAirdrops := []*types.Airdrop{}
	for _, airdrop := range params.Airdrops {
		if airdrop.AirdropIdentifier == identifier {
			airdrop.ClaimedSoFar = airdrop.ClaimedSoFar.Add(amount)
		}
		newAirdrops = append(newAirdrops, airdrop)
	}
	params.Airdrops = newAirdrops
	return k.SetParams(ctx, params)
}

// ResetClaimedSoFar resets ClaimedSoFar for a all airdrops
func (k Keeper) ResetClaimedSoFar(ctx sdk.Context, airdropIdentifier string) error {
	params, err := k.GetParams(ctx)
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] params.Airdrops %v", params.Airdrops))
	if err != nil {
		panic(err)
	}

	newAirdrops := []*types.Airdrop{}
	for _, airdrop := range params.Airdrops {
		if airdrop.AirdropIdentifier == airdropIdentifier {
			airdrop.ClaimedSoFar = sdkmath.ZeroInt()
			k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] resetting claimSoFar for %s", airdropIdentifier))
		}
		newAirdrops = append(newAirdrops, airdrop)
	}
	params.Airdrops = newAirdrops
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] params.Airdrops %v", params.Airdrops))
	return k.SetParams(ctx, params)
}

// DeleteAirdropAndEpoch deletes existing airdrop and corresponding epoch.
func (k Keeper) DeleteAirdropAndEpoch(ctx sdk.Context, identifier string) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	newAirdrops := []*types.Airdrop{}
	for _, airdrop := range params.Airdrops {
		if airdrop.AirdropIdentifier != identifier {
			newAirdrops = append(newAirdrops, airdrop)
		}
	}
	params.Airdrops = newAirdrops
	k.epochsKeeper.DeleteEpochInfo(ctx, fmt.Sprintf("airdrop-%s", identifier))
	return k.SetParams(ctx, params)
}

func (k Keeper) UpdateAirdropAddress(ctx sdk.Context, existingStrideAddress string, newStrideAddress string, airdropId string) error {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropId)
	if airdrop == nil {
		return errorsmod.Wrapf(types.ErrAirdropNotFound, fmt.Sprintf("airdrop not found for identifier %s", airdropId))
	}

	// verify that the strideAddress is valid
	_, err := sdk.AccAddressFromBech32(newStrideAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid stride address %s", newStrideAddress))
	}

	// note: existingAccAddress will be a STRIDE address with the same coin type as existingAddress
	// when new airdrops are ingested, we call utils.ConvertAddressToStrideAddress to convert
	// the host zone (e.g. Evmos) address to a Stride address. The same conversion must be done
	// if you're attempting to access a claim record for a non-Stride-address.
	existingAccAddress, err := sdk.AccAddressFromBech32(existingStrideAddress)
	if err != nil {
		return errorsmod.Wrapf(types.ErrClaimNotFound,
			fmt.Sprintf("error getting claim record for address %s on airdrop %s", existingStrideAddress, airdropId))
	}
	claimRecord, err := k.GetClaimRecord(ctx, existingAccAddress, airdrop.AirdropIdentifier)
	if (err != nil) || (claimRecord.Address == "") {
		return errorsmod.Wrapf(types.ErrClaimNotFound,
			fmt.Sprintf("error getting claim record for address %s on airdrop %s", existingStrideAddress, airdropId))
	}

	claimRecord.Address = newStrideAddress
	err = k.SetClaimRecord(ctx, claimRecord) // this does NOT delete the old record, because claims are indexed by address
	if err != nil {
		return errorsmod.Wrapf(types.ErrModifyingClaimRecord,
			fmt.Sprintf("error setting claim record from address %s to address %s on airdrop %s", existingStrideAddress, newStrideAddress, airdropId))
	}

	// this deletes the old record
	err = k.DeleteClaimRecord(ctx, existingAccAddress, airdrop.AirdropIdentifier)
	if err != nil {
		return errorsmod.Wrapf(types.ErrModifyingClaimRecord,
			fmt.Sprintf("error deleting claim record for address %s on airdrop %s", existingStrideAddress, airdropId))
	}

	return nil
}
