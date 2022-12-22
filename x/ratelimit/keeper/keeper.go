package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

type (
	Keeper struct {
		storeKey   storetypes.StoreKey
		cdc        codec.BinaryCodec
		paramstore paramtypes.Subspace

		bankKeeper    types.BankKeeper
		channelKeeper channelkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	channelKeeper channelkeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      key,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		channelKeeper: channelKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
