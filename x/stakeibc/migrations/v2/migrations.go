package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oldstakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func migrateHostZone(store sdk.KVStore, cdc codec.BinaryCodec) error {
	paramsStore := prefix.NewStore(store, []byte(stakeibctypes.HostZoneKey))

	iterator := paramsStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldHostZone oldstakeibctypes.HostZone
		err := cdc.Unmarshal(iterator.Value(), &oldHostZone)
		if err != nil {
			return err
		}

		// Convert and serialize using the new type
		newHostZone := convertToNewHostZone(oldHostZone)
		newHostZoneBz, err := cdc.Marshal(&newHostZone)
		if err != nil {
			return err
		}

		// Store new type
		paramsStore.Set(iterator.Key(), newHostZoneBz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	return migrateHostZone(store, cdc)
}
