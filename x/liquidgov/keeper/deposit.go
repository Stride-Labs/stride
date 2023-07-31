package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
)

// GetProposal gets a proposal from store by ProposalID.
// Panics if can't unmarshal the proposal.
func (keeper Keeper) GetDeposit(ctx sdk.Context, creator string, hostZoneId string) (types.Deposit, bool) {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.DepositKey(creator, hostZoneId))
	if bz == nil {
		return types.Deposit{}, false
	}

	var deposit types.Deposit
	if err := keeper.UnmarshalDeposit(bz, &deposit); err != nil {
		panic(err)
	}

	return deposit, true
}

// SetDeposit sets a proposal to store.
// Panics if can't marshal the deposit.
func (keeper Keeper) SetDeposit(ctx sdk.Context, deposit types.Deposit) {
	bz, err := keeper.MarshalDeposit(deposit)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.DepositKey(deposit.Creator, deposit.HostZoneId), bz)
}

// DeleteDeposit deletes a deposit from store.
// Panics if the deposit doesn't exist.
func (keeper Keeper) DeleteDeposit(ctx sdk.Context, creator string, hostZoneId string) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(types.DepositKey(creator, hostZoneId))
}

// IterateDeposits iterates over the all the deposits and performs a callback function.
// Panics when the iterator encounters a deposit which can't be unmarshaled.
func (keeper Keeper) IterateDeposits(ctx sdk.Context, cb func(deposit types.Deposit) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.DepositsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		err := keeper.UnmarshalDeposit(iterator.Value(), &deposit)
		if err != nil {
			panic(err)
		}

		if cb(deposit) {
			break
		}
	}
}

// GetDeposits returns all the deposits from store
func (keeper Keeper) GetDeposits(ctx sdk.Context) (deposits []types.Deposit) {
	keeper.IterateDeposits(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}

// Helper function to iterate votes and see how much of deposit is available now
func (keeper Keeper) DepositAvailableNow(ctx sdk.Context, creator string, hostZoneId string) (sdk.Int) {
	deposit, _ := keeper.GetDeposit(ctx, creator, hostZoneId)
	keeper.Logger(ctx).Info(fmt.Sprintf("looking for available deposit creator:%s hostZone:%s deposit:%v", creator, hostZoneId, deposit))
	return deposit.Amount // Check with when votes happened/proposals end in future...
}

func (keeper Keeper) MarshalDeposit(deposit types.Deposit) ([]byte, error) {
	bz, err := deposit.Marshal()
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (keeper Keeper) UnmarshalDeposit(bz []byte, deposit *types.Deposit) error {
	err := deposit.Unmarshal(bz)
	if err != nil {
		return err
	}
	return nil
}
