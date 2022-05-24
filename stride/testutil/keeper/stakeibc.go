package keeper

import (
	"testing"

	"github.com/Stride-Labs/cosmos-sdk/codec"
	codectypes "github.com/Stride-Labs/cosmos-sdk/codec/types"
	"github.com/Stride-Labs/cosmos-sdk/store"
	storetypes "github.com/Stride-Labs/cosmos-sdk/store/types"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	capabilitykeeper "github.com/Stride-Labs/cosmos-sdk/x/capability/keeper"
	typesparams "github.com/Stride-Labs/cosmos-sdk/x/params/types"
	ibckeeper "github.com/Stride-Labs/ibc-go/v3/modules/core/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func StakeibcKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	logger := log.NewNopLogger()

	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(registry)
	capabilityKeeper := capabilitykeeper.NewKeeper(appCodec, storeKey, memStoreKey)

	ss := typesparams.NewSubspace(appCodec,
		types.Amino,
		storeKey,
		memStoreKey,
		"StakeibcSubSpace",
	)
	IBCKeeper := ibckeeper.NewKeeper(
		appCodec,
		storeKey,
		ss,
		nil,
		nil,
		capabilityKeeper.ScopeToModule("StakeibcIBCKeeper"),
	)

	paramsSubspace := typesparams.NewSubspace(appCodec,
		types.Amino,
		storeKey,
		memStoreKey,
		"StakeibcParams",
	)
	k := keeper.NewKeeper(
		appCodec,
		storeKey,
		memStoreKey,
		paramsSubspace,
		// IBCKeeper.ChannelKeeper,
		// &IBCKeeper.PortKeeper,
		// capabilityKeeper.ScopeToModule("StakeibcScopedKeeper"),
		IBCKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
