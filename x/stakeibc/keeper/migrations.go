package keeper

import (
	// v2 "github.com/Stride-Labs/stride/x/stakeibc/migrations/v2"

	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/store/prefix"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	stakeIbcStoreKey := m.keeper.storeKey
	stakeIbcStore := prefix.NewStore(ctx.KVStore(stakeIbcStoreKey), stakeibctypes.KeyPrefix(stakeibctypes.HostZoneKey))

	iterator := sdk.KVStorePrefixIterator(stakeIbcStore, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var hostZone stakeibctypes.HostZone
		m.keeper.cdc.MustUnmarshal(iterator.Value(), &hostZone)

		if hostZone.ChainId == "GAIA" {
			hostZone.Bech32Prefix = "cosmos"
			m.keeper.SetHostZone(ctx, hostZone)
		}
	}

	return nil

	// return v2.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}
