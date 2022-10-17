package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/gogo/protobuf/proto"

	"github.com/Stride-Labs/stride/utils"
	"github.com/Stride-Labs/stride/x/claim/types"
	vestingtypes "github.com/Stride-Labs/stride/x/claim/vesting/types"
)

func (k Keeper) LoadAllocationData(ctx sdk.Context, allocationData string) bool {
	totalWeight := sdk.NewDec(0)
	records := []types.ClaimRecord{}

	lines := strings.Split(allocationData, "\n")
	allocatedFlags := map[string]bool{}
	for _, line := range lines {
		data := strings.Split(line, ",")
		if data[0] == "" || data[1] == "" {
			continue
		}

		weight, err := sdk.NewDecFromStr(data[1])
		if err != nil || allocatedFlags[data[0]] {
			continue
		}

		_, err = sdk.AccAddressFromBech32(data[0])
		if err != nil {
			continue
		}

		records = append(records, types.ClaimRecord{
			Address:         data[0],
			Weight:          weight,
			ActionCompleted: []bool{false, false, false},
		})

		totalWeight = totalWeight.Add(weight)
		allocatedFlags[data[0]] = true
	}

	k.SetTotalWeight(ctx, totalWeight)
	k.SetClaimRecords(ctx, records)
	return true
}

// GetModuleAccountAddress gets the module account address
func (k Keeper) GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

// GetModuleAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountBalance(ctx sdk.Context) sdk.Coin {
	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return k.bankKeeper.GetBalance(ctx, moduleAccAddr, params.ClaimDenom)
}

func (k Keeper) EndAirdrop(ctx sdk.Context) error {
	ctx.Logger().Info("Beginning EndAirdrop logic")
	err := k.fundRemainingsToCommunity(ctx)
	if err != nil {
		return err
	}
	ctx.Logger().Info("Clearing claims module state entries")
	k.clearInitialClaimables(ctx)

	ctx.Logger().Info("Beginning clawback")
	err = k.ClawbackAirdrop(ctx)
	if err != nil {
		return err
	}
	return nil
}

// SweepAirdrop sweep all airdrop rewards back into the airdrop distribution account
func (k Keeper) SweepAirdrop(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(params.DistributorAddress)
	if err != nil {
		return err
	}

	bal := k.bankKeeper.GetBalance(ctx, addr, params.ClaimDenom)
	sweepCoins := sdk.NewCoins(bal)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sweepCoins)
	if err != nil {
		return err
	}
	return nil
}

// ClawbackAirdrop claws back all the Stride coins from airdrop
func (k Keeper) ClawbackAirdrop(ctx sdk.Context) error {
	addr := k.GetModuleAccountAddress(ctx)
	bal := k.GetModuleAccountBalance(ctx)

	totalClawback := sdk.NewCoins(bal)
	err := k.distrKeeper.FundCommunityPool(ctx, totalClawback, addr)
	if err != nil {
		return err
	}

	ctx.Logger().Info(fmt.Sprintf("clawed back %d ustrd into community pool", totalClawback.AmountOf("ustrd").Int64()))
	return nil
}

// ClearClaimables clear claimable amounts
func (k Keeper) clearInitialClaimables(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.ClaimRecordsStorePrefix))
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
func (k Keeper) GetClaimRecords(ctx sdk.Context) []types.ClaimRecord {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimRecordsStorePrefix))

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
func (k Keeper) GetClaimRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimRecord, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimRecordsStorePrefix))
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
func (k Keeper) SetTotalWeight(ctx sdk.Context, totalWeight sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.TotalWeightKey), []byte(totalWeight.String()))
}

// GetTotalWeight gets total sum of user weights in store
func (k Keeper) GetTotalWeight(ctx sdk.Context) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(types.TotalWeightKey))
	if b == nil {
		return sdk.ZeroDec(), types.ErrTotalWeightNotSet
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
	prefixStore := prefix.NewStore(store, []byte(types.ClaimRecordsStorePrefix))

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
func (k Keeper) GetAirdropDistributor(ctx sdk.Context) (sdk.AccAddress, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddressFromBech32(params.DistributorAddress)
}

// Get airdrop claim denom
func (k Keeper) GetAirdropClaimDenom(ctx sdk.Context) (string, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", err
	}

	return params.ClaimDenom, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error) {
	claimRecord, err := k.GetClaimRecord(ctx, addr)
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

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// If we are before the start time, do nothing.
	// This case _shouldn't_ occur on chain, since the
	// start time ought to be chain start time.
	if ctx.BlockTime().Before(params.AirdropStartTime) {
		return sdk.Coins{}, nil
	}

	totalWeight, err := k.GetTotalWeight(ctx)
	if err != nil {
		return nil, types.ErrFailedToGetTotalWeight
	}

	poolBal := k.GetModuleAccountBalance(ctx)

	percentageForAction := types.PercentageForFree
	if action == types.ActionDelegateStake {
		percentageForAction = types.PercentageForStake
	} else if action == types.ActionLiquidStake {
		percentageForAction = types.PercentageForLiquidStake
	}

	claimableAmount := poolBal.Amount.ToDec().
		Mul(percentageForAction).
		Mul(claimRecord.Weight).
		Quo(totalWeight).RoundInt()
	claimableCoins := sdk.NewCoins(sdk.NewCoin(params.ClaimDenom, claimableAmount))

	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)
	// The entire airdrop has completed
	if elapsedAirdropTime > params.AirdropDuration {
		return sdk.Coins{}, nil
	}
	return claimableCoins, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	claimRecord, err := k.GetClaimRecord(ctx, addr)
	if err != nil {
		return sdk.Coins{}, err
	}
	if claimRecord.Address == "" {
		return sdk.Coins{}, nil
	}

	totalClaimable := sdk.Coins{}

	for action := range types.Action_name {
		claimableForAction, err := k.GetClaimableAmountForAction(ctx, addr, types.Action(action))
		if err != nil {
			return sdk.Coins{}, err
		}
		if !claimableForAction.Empty() {
			totalClaimable = totalClaimable.Add(claimableForAction...)
		}
	}
	return totalClaimable, nil
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error) {
	claimableAmount, err := k.GetClaimableAmountForAction(ctx, addr, action)
	if err != nil {
		return claimableAmount, err
	}

	if claimableAmount.Empty() {
		return claimableAmount, nil
	}

	claimRecord, err := k.GetClaimRecord(ctx, addr)
	if err != nil {
		return nil, err
	}

	// Check if vesting tokens already exist for this account.
	if claimRecord.ActionCompleted[types.ActionDelegateStake] || claimRecord.ActionCompleted[types.ActionLiquidStake] {
		// Get user stride vesting account and grant a new vesting
		acc := k.accountKeeper.GetAccount(ctx, addr)
		vestingAcc, isVesting := acc.(*vestingtypes.StridePeriodicVestingAccount)
		if !isVesting {
			return nil, err
		}

		periodLength := utils.GetAirdropDurationForAction(action)
		vestingAcc.AddNewGrant(vestingtypes.Period{
			StartTime: ctx.BlockTime().Unix(),
			Length:    periodLength,
			Amount:    claimableAmount,
		})

		k.accountKeeper.SetAccount(ctx, vestingAcc)
	} else {
		// If the account is a default vesting account, don't convert it to stride vesting account.
		acc := k.accountKeeper.GetAccount(ctx, addr)
		_, isDefaultVestingAccount := acc.(*authvestingtypes.BaseVestingAccount)
		if isDefaultVestingAccount {
			return nil, err
		}

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
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimableAmount)
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

// FundRemainingsToCommunity fund remainings to the community when airdrop period end
func (k Keeper) fundRemainingsToCommunity(ctx sdk.Context) error {
	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	amt := k.GetModuleAccountBalance(ctx)
	ctx.Logger().Info(fmt.Sprintf(
		"Sending %d %s to community pool, corresponding to the 'unclaimed airdrop'", amt.Amount.Int64(), amt.Denom))
	return k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(amt), moduleAccAddr)
}
