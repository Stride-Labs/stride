package keeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/gogo/protobuf/proto"

	"github.com/Stride-Labs/stride/utils"
	"github.com/Stride-Labs/stride/x/claim/types"
	vestingtypes "github.com/Stride-Labs/stride/x/claim/vesting/types"
	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
)

func (k Keeper) LoadAllocationData(ctx sdk.Context, allocationData string) bool {
	totalWeight := sdk.NewDec(0)
	records := []types.ClaimRecord{}
	airdropIdentifier := ""

	lines := strings.Split(allocationData, "\n")
	allocatedFlags := map[string]bool{}
	for _, line := range lines {
		data := strings.Split(line, ",")
		if data[0] == "" || data[1] == "" || data[2] == "" {
			continue
		}

		weight, err := sdk.NewDecFromStr(data[2])
		if err != nil || allocatedFlags[data[1]] {
			continue
		}

		_, err = sdk.AccAddressFromBech32(data[1])
		if err != nil {
			continue
		}

		records = append(records, types.ClaimRecord{
			AirdropIdentifier: data[0],
			Address:           data[1],
			Weight:            weight,
			ActionCompleted:   []bool{false, false, false},
		})

		totalWeight = totalWeight.Add(weight)
		allocatedFlags[data[1]] = true
		airdropIdentifier = data[0]
	}

	k.SetTotalWeight(ctx, totalWeight, airdropIdentifier)
	k.SetClaimRecords(ctx, records)
	return true
}

func (k Keeper) RemoveDuplicatedAirdrops(ctx sdk.Context, identifier string, users []string, weights []sdk.Dec) ([]string, []sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ClaimRecordsStorePrefix), []byte(identifier)...))
	newUsers := []string{}
	newWeights := []sdk.Dec{}
	for idx, user := range users {
		addr, _ := sdk.AccAddressFromBech32(user)
		// If new user, then append user and weight
		if !prefixStore.Has(addr) {
			newUsers = append(newUsers, user)
			newWeights = append(newWeights, weights[idx])
		}
	}

	return newUsers, newWeights
}

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

// GetDistributorAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetDistributorAccountBalance(ctx sdk.Context, airdropIdentifier string) (sdk.Coin, error) {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return sdk.Coin{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid airdrop identifier: GetDistributorAccountBalance")
	}

	addr, err := k.GetAirdropDistributor(ctx, airdropIdentifier)
	if err != nil {
		return sdk.Coin{}, err
	}
	return k.bankKeeper.GetBalance(ctx, addr, airdrop.ClaimDenom), nil
}

func (k Keeper) EndAirdrop(ctx sdk.Context, airdropIdentifier string) error {
	ctx.Logger().Info("Clearing claims module state entries")
	k.clearInitialClaimables(ctx, airdropIdentifier)
	k.SetTotalWeight(ctx, sdk.ZeroDec(), airdropIdentifier)
	k.DeleteAirdropAndEpoch(ctx, airdropIdentifier)
	return nil
}

// // ClawbackAirdrop claws back all the Stride coins from airdrop
// func (k Keeper) ClawbackAirdrop(ctx sdk.Context, airdropIdentifier string) error {
// 	addr, err := k.GetAirdropDistributor(ctx, airdropIdentifier)
// 	bal := k.GetDistributorAccountBalance(ctx, airdropIdentifier)

// 	totalClawback := sdk.NewCoins(bal)
// 	err = k.distrKeeper.FundCommunityPool(ctx, totalClawback, addr)
// 	if err != nil {
// 		return err
// 	}

// 	ctx.Logger().Info(fmt.Sprintf("clawed back %d ustrd into community pool", totalClawback.AmountOf("ustrd").Int64()))
// 	return nil
// }

// ClearClaimedStatus clear users' claimed status
func (k Keeper) ClearClaimedStatus(ctx sdk.Context, airdropIdentifier string) {
	records := k.GetClaimRecords(ctx, airdropIdentifier)
	for idx := range records {
		records[idx].ActionCompleted = []bool{false, false, false}
	}

	k.SetClaimRecords(ctx, records)
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

// SetClaimables set claimable amount from balances object
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

// Get airdrop distributor address
func (k Keeper) GetAirdropDistributor(ctx sdk.Context, airdropIdentifier string) (sdk.AccAddress, error) {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return sdk.AccAddress{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid airdrop identifier: GetAirdropDistributor")
	}
	return sdk.AccAddressFromBech32(airdrop.DistributorAddress)
}

// Get airdrop claim denom
func (k Keeper) GetAirdropClaimDenom(ctx sdk.Context, airdropIdentifier string) (string, error) {
	airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
	if airdrop == nil {
		return "", sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid airdrop identifier: GetAirdropClaimDenom")
	}
	return airdrop.ClaimDenom, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action, airdropIdentifier string) (sdk.Coins, error) {
	claimRecord, err := k.GetClaimRecord(ctx, addr, airdropIdentifier)
	if err != nil {
		return nil, err
	}

	if claimRecord.Address == "" {
		return sdk.Coins{}, nil
	}

	// if action already completed, nothing is claimable
	if claimRecord.ActionCompleted[action] {
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
	if action == types.ActionDelegateStake {
		percentageForAction = types.PercentageForStake
	} else if action == types.ActionLiquidStake {
		percentageForAction = types.PercentageForLiquidStake
	}

	poolBal, err := k.GetDistributorAccountBalance(ctx, airdropIdentifier)
	if err != nil {
		return sdk.Coins{}, err
	}

	claimableAmount := poolBal.Amount.ToDec().
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

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress, airdropIdentifier string) (sdk.Coins, error) {
	claimRecord, err := k.GetClaimRecord(ctx, addr, airdropIdentifier)
	if err != nil {
		return sdk.Coins{}, err
	}
	if claimRecord.Address == "" {
		return sdk.Coins{}, nil
	}

	totalClaimable := sdk.Coins{}

	for action := range types.Action_name {
		claimableForAction, err := k.GetClaimableAmountForAction(ctx, addr, types.Action(action), airdropIdentifier)
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
func (k Keeper) GetAirdropIdentifierForUser(ctx sdk.Context, addr sdk.AccAddress) string {
	records := k.GetClaimRecords(ctx, "")
	for _, record := range records {
		if record.Address == addr.String() {
			return record.AirdropIdentifier
		}
	}
	return ""
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action, airdropIdentifier string) (sdk.Coins, error) {
	if airdropIdentifier == "" {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid airdrop identifier: ClaimCoinsForAction")
	}

	claimableAmount, err := k.GetClaimableAmountForAction(ctx, addr, action, airdropIdentifier)
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

	// If the account is a default vesting account, don't grant new vestings
	acc := k.accountKeeper.GetAccount(ctx, addr)
	_, isDefaultVestingAccount := acc.(*authvestingtypes.BaseVestingAccount)
	if isDefaultVestingAccount {
		return nil, err
	}

	acc = k.accountKeeper.GetAccount(ctx, addr)
	strideVestingAcc, isStrideVestingAccount := acc.(*vestingtypes.StridePeriodicVestingAccount)
	// Check if vesting tokens already exist for this account.
	if !isStrideVestingAccount {
		// Convert user account into stride veting account.
		baseAccount := k.accountKeeper.NewAccountWithAddress(ctx, addr)
		if _, ok := baseAccount.(*authtypes.BaseAccount); !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid account type; expected: BaseAccount, got: %T", baseAccount)
		}

		periodLength := utils.GetAirdropDurationForAction(action)
		vestingAcc := vestingtypes.NewStridePeriodicVestingAccount(baseAccount.(*authtypes.BaseAccount), claimableAmount, []vestingtypes.Period{{
			StartTime: ctx.BlockTime().Unix(),
			Length:    periodLength,
			Amount:    claimableAmount,
		}})
		k.accountKeeper.SetAccount(ctx, vestingAcc)
	} else {
		// Grant a new vesting to the existing stride vesting account
		periodLength := utils.GetAirdropDurationForAction(action)
		strideVestingAcc.AddNewGrant(vestingtypes.Period{
			StartTime: ctx.BlockTime().Unix(),
			Length:    periodLength,
			Amount:    claimableAmount,
		})
		k.accountKeeper.SetAccount(ctx, strideVestingAcc)
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
func (k Keeper) CreateAirdropAndEpoch(ctx sdk.Context, distributor string, denom string, startTime uint64, duration uint64, identifier string) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	for _, airdrop := range params.Airdrops {
		if airdrop.AirdropIdentifier == identifier {
			return types.ErrAirdropAlreadyExists
		}
	}

	airdrop := types.Airdrop{
		AirdropIdentifier:  identifier,
		AirdropDuration:    time.Duration(duration * uint64(time.Second)),
		ClaimDenom:         denom,
		DistributorAddress: distributor,
		AirdropStartTime:   time.Unix(int64(startTime), 0),
	}

	params.Airdrops = append(params.Airdrops, &airdrop)
	k.epochsKeeper.SetEpochInfo(ctx, epochstypes.EpochInfo{
		Identifier:              fmt.Sprintf("airdrop-%s", identifier),
		StartTime:               airdrop.AirdropStartTime.Add(time.Minute),
		Duration:                time.Hour * 24 * 30,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	})
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
