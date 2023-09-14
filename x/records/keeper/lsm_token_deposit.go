package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/records/types"
)

func (k Keeper) SetLSMTokenDeposit(ctx sdk.Context, deposit types.LSMTokenDeposit) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	depositKey := types.GetLSMTokenDepositKey(deposit.ChainId, deposit.Denom)
	depositData := k.Cdc.MustMarshal(&deposit)
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
	k.Cdc.MustUnmarshal(depositData, &deposit)
	return deposit, true
}

func (k Keeper) GetAllLSMTokenDeposit(ctx sdk.Context) []types.LSMTokenDeposit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	iterator := store.Iterator(nil, nil)
	allLSMTokenDeposits := []types.LSMTokenDeposit{}

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.LSMTokenDeposit
		k.Cdc.MustUnmarshal(iterator.Value(), &deposit)
		allLSMTokenDeposits = append(allLSMTokenDeposits, deposit)
	}

	return allLSMTokenDeposits
}

func (k Keeper) UpdateLSMTokenDepositStatus(ctx sdk.Context, deposit types.LSMTokenDeposit, status types.LSMTokenDeposit_Status) {
	deposit.Status = status
	k.SetLSMTokenDeposit(ctx, deposit)
}

func (k Keeper) GetLSMDepositsForHostZone(ctx sdk.Context, chainId string) []types.LSMTokenDeposit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LSMTokenDepositKey))
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(chainId))
	hostZoneLSMTokenDeposits := []types.LSMTokenDeposit{}

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.LSMTokenDeposit
		k.Cdc.MustUnmarshal(iterator.Value(), &deposit)
		hostZoneLSMTokenDeposits = append(hostZoneLSMTokenDeposits, deposit)
	}

	return hostZoneLSMTokenDeposits
}

func (k Keeper) GetLSMDepositsForHostZoneWithStatus(ctx sdk.Context, chainId string, status types.LSMTokenDeposit_Status) []types.LSMTokenDeposit {
	filtered := []types.LSMTokenDeposit{}
	hostZoneLSMTokenDeposits := k.GetLSMDepositsForHostZone(ctx, chainId)
	for _, deposit := range hostZoneLSMTokenDeposits {
		if deposit.Status == status {
			filtered = append(filtered, deposit)
		}
	}
	return filtered
}
