package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	icqkeeper "github.com/Stride-Labs/stride/v5/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		stakeibcKeeper        types.StakeibcKeeper
		accountKeeper         types.AccountKeeper
		bankKeeper            types.BankKeeper
		InterchainQueryKeeper icqkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	stakeibcKeeper types.StakeibcKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	interchainQueryKeeper icqkeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		stakeibcKeeper:        stakeibcKeeper,
		accountKeeper:         accountKeeper,
		bankKeeper:            bankKeeper,
		InterchainQueryKeeper: interchainQueryKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
