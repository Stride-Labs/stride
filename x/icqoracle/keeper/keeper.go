package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

type Keeper struct {
	cdc               codec.Codec
	storeKey          storetypes.StoreKey
	IcqKeeper         types.IcqKeeper
	ibcTransferKeeper types.IbcTransferKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	icqKeeper types.IcqKeeper,
	ibcTransferKeeper types.IbcTransferKeeper,
) *Keeper {
	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		IcqKeeper:         icqKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
