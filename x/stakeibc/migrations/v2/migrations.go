// package v2

// import (
// 	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
// 	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
// 	"github.com/cosmos/cosmos-sdk/codec"
// 	"github.com/cosmos/cosmos-sdk/store/prefix"
// 	storetypes "github.com/cosmos/cosmos-sdk/store/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// func MigrateStore(
// 	ctx sdk.Context,
// 	storeKey storetypes.StoreKey,
// 	keeper stakeibckeeper.Keeper,
// 	cdc codec.BinaryCodec,
// ) error {
// 	store := prefix.NewStore(ctx.KVStore(storeKey), stakeibctypes.KeyPrefix(stakeibctypes.HostZoneKey))

// 	iterator := sdk.KVStorePrefixIterator(store, []byte{})
// 	defer iterator.Close()

// 	for ; iterator.Valid(); iterator.Next() {
// 		var hostZone stakeibctypes.HostZone
// 		cdc.MustUnmarshal(iterator.Value(), &hostZone)

// 		if hostZone.ChainId == "GAIA" {
// 			hostZone.Bech32Prefix = "cosmos"
// 			keeper.SetHostZone(ctx, hostZone)
// 		}
// 	}

// 	return nil
// }
