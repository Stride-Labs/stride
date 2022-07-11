package keeper

import (
	// v2 "github.com/Stride-Labs/stride/x/stakeibc/migrations/v2"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	// channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
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

	ibcChannelStoreKey := storetypes.NewKVStoreKey(ibcchanneltypes.StoreKey)
	ibcChannelKeeper := m.keeper.IBCKeeper.ChannelKeeper

	ibcChannelStore := ctx.KVStore(ibcChannelStoreKey)

	for _, packet := range ibcChannelKeeper.GetAllPacketCommitments(ctx) {
		packetKey := host.PacketCommitmentKey(packet.PortId, packet.ChannelId, packet.Sequence)
		ibcChannelStore.Delete(packetKey)
	}

	return nil

	// return v2.MigrateStore(ctx, stakeIbcStoreKey, ibcChannelStoreKey, ibcChannelKeeper)
}
