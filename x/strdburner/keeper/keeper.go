package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetStrdBurnerAddress() sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

func (k Keeper) SetTotalStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	bz := sdk.Uint64ToBigEndian(amount.Uint64())
	ctx.KVStore(k.storeKey).Set([]byte(types.TotalStrdBurnedKey), bz)
}

func (k Keeper) GetTotalStrdBurned(ctx sdk.Context) sdkmath.Int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(types.TotalStrdBurnedKey))

	// If no value has been set, return zero
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(bz))
}

func (k Keeper) SetProtocolStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	bz := sdk.Uint64ToBigEndian(amount.Uint64())
	ctx.KVStore(k.storeKey).Set([]byte(types.ProtocolStrdBurnedKey), bz)
}

func (k Keeper) GetProtocolStrdBurned(ctx sdk.Context) sdkmath.Int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(types.ProtocolStrdBurnedKey))

	// If no value has been set, return zero
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(bz))
}

func (k Keeper) SetUserStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	bz := sdk.Uint64ToBigEndian(amount.Uint64())
	ctx.KVStore(k.storeKey).Set([]byte(types.UserStrdBurnedKey), bz)
}

func (k Keeper) GetUserStrdBurned(ctx sdk.Context) sdkmath.Int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(types.UserStrdBurnedKey))

	// If no value has been set, return zero
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(bz))
}
