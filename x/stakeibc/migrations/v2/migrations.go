package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldstakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func migrateHostZone(store sdk.KVStore, cdc codec.BinaryCodec) error {
	stakeibcStore := prefix.NewStore(store, []byte(stakeibctypes.HostZoneKey))

	iterator := stakeibcStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldHostZone oldstakeibctypes.HostZone
		err := cdc.Unmarshal(iterator.Value(), &oldHostZone)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal host zone (%v) using old data type", iterator.Key())
		}

		// Convert and serialize using the new type
		newHostZone := convertToNewHostZone(oldHostZone)
		newHostZoneBz, err := cdc.Marshal(&newHostZone)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal host zone (%v) using new data type", iterator.Key())
		}

		// Store new type
		stakeibcStore.Set(iterator.Key(), newHostZoneBz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	return migrateHostZone(store, cdc)
}
