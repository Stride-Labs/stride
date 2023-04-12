package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) SetLSMTokenDeposit(ctx sdk.Context, deposit types.LSMTokenDeposit) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	depositKey := types.GetLSMTokenDepositKey(deposit.ChainId, deposit.Denom)
	depositData := k.cdc.MustMarshal(&deposit)
	store.Set(depositKey, depositData)
}

func (k Keeper) RemoveLSMTokenDeposit(ctx sdk.Context, chainId, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	depositKey := types.GetLSMTokenDepositKey(chainId, denom)
	store.Delete(depositKey)
}

func (k Keeper) GetLSMTokenDeposit(ctx sdk.Context, chainId, denom string) (deposit types.LSMTokenDeposit, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	depositKey := types.GetLSMTokenDepositKey(chainId, denom)
	depositData := store.Get(depositKey)
	if len(depositData) == 0 {
		return deposit, false
	}
	k.cdc.MustUnmarshal(depositData, &deposit)
	return deposit, true
}

func (k Keeper) GetAllLSMTokenDeposit(ctx sdk.Context) []types.LSMTokenDeposit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	iterator := store.Iterator(nil, nil)
	allLSMTokenDeposits := []types.LSMTokenDeposit{}

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.LSMTokenDeposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		allLSMTokenDeposits = append(allLSMTokenDeposits, deposit)
	}

	return allLSMTokenDeposits
}

func (k Keeper) AddLSMTokenDeposit(ctx sdk.Context, deposit types.LSMTokenDeposit) {
	// TODO [LSM]
	// See if a deposit already exists for this denom
	// If so, increment the amount
	// otherwise, create a new deposit
}

func (k Keeper) UpdateLSMTokenDepositStatus(ctx sdk.Context, deposit types.LSMTokenDeposit, status types.LSMDepositStatus) {
	// TODO [LSM]
}

func (k Keeper) GetLSMDepositsForHostZone(ctx sdk.Context, deposit types.LSMTokenDeposit, chainId string) []types.LSMTokenDeposit {
	// TODO [LSM]
	return []types.LSMTokenDeposit{}
}

func (k Keeper) GetLSMDepositsForHostZoneWithStatus(
	ctx sdk.Context,
	deposit types.LSMTokenDeposit,
	chainId string,
	status types.LSMDepositStatus,
) []types.LSMTokenDeposit {
	// TODO [LSM]
	return []types.LSMTokenDeposit{}
}
