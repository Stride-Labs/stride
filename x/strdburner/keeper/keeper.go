package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/strdburner/types"
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

func (k Keeper) SetTotalStrdBurned(ctx sdk.Context, amount math.Int) {
	bz := sdk.Uint64ToBigEndian(amount.Uint64())
	ctx.KVStore(k.storeKey).Set([]byte(types.TotalStrdBurnedKey), bz)
}

func (k Keeper) GetTotalStrdBurned(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(types.TotalStrdBurnedKey))

	// If no value has been set, return zero
	if bz == nil {
		return math.ZeroInt()
	}

	return math.NewIntFromUint64(sdk.BigEndianToUint64(bz))
}
