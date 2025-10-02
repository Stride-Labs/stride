package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"cosmossdk.io/store/prefix"
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

// Sets the total STRD burned from the protocol
func (k Keeper) SetProtocolStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	bz := sdk.Uint64ToBigEndian(amount.Uint64())
	ctx.KVStore(k.storeKey).Set([]byte(types.ProtocolStrdBurnedKey), bz)
}

// Returns the total STRD burned from the protocol
func (k Keeper) GetProtocolStrdBurned(ctx sdk.Context) sdkmath.Int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(types.ProtocolStrdBurnedKey))

	// If no value has been set, return zero
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(bz))
}

// Sets the total STRD burned from all users
func (k Keeper) SetTotalUserStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	bz := sdk.Uint64ToBigEndian(amount.Uint64())
	ctx.KVStore(k.storeKey).Set([]byte(types.TotalUserStrdBurnedKey), bz)
}

// Returns the total STRD burned from all users
func (k Keeper) GetTotalUserStrdBurned(ctx sdk.Context) sdkmath.Int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(types.TotalUserStrdBurnedKey))

	// If no value has been set, return zero
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(bz))
}

// Sets the total STRD burned from a given address
func (k Keeper) SetStrdBurnedByAddress(ctx sdk.Context, address sdk.AccAddress, amount sdkmath.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BurnedByAddressKeyPrefix)
	amountBz := sdk.Uint64ToBigEndian(amount.Uint64())
	store.Set(address, amountBz)
}

// Returns the total STRD burned from a given address
func (k Keeper) GetStrdBurnedByAddress(ctx sdk.Context, address sdk.AccAddress) sdkmath.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BurnedByAddressKeyPrefix)

	amountBz := store.Get(address)
	if len(amountBz) == 0 {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(amountBz))
}

// Returns the total STRD burned across all users and the protocol
func (k Keeper) GetTotalStrdBurned(ctx sdk.Context) sdkmath.Int {
	protocolBurned := k.GetProtocolStrdBurned(ctx)
	userBurned := k.GetTotalUserStrdBurned(ctx)
	return protocolBurned.Add(userBurned)
}

// Increment the protocol strd burned
func (k Keeper) IncrementProtocolStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	currentBurned := k.GetProtocolStrdBurned(ctx)
	newBurned := currentBurned.Add(amount)
	k.SetProtocolStrdBurned(ctx, newBurned)
}

// Increment the total user strd burned
func (k Keeper) IncrementTotalUserStrdBurned(ctx sdk.Context, amount sdkmath.Int) {
	currentBurned := k.GetTotalUserStrdBurned(ctx)
	newBurned := currentBurned.Add(amount)
	k.SetTotalUserStrdBurned(ctx, newBurned)
}

// Increment the strd burned for an address
func (k Keeper) IncrementStrdBurnedByAddress(ctx sdk.Context, address sdk.AccAddress, amount sdkmath.Int) {
	currentBurned := k.GetStrdBurnedByAddress(ctx, address)
	newBurned := currentBurned.Add(amount)
	k.SetStrdBurnedByAddress(ctx, address, newBurned)
}
