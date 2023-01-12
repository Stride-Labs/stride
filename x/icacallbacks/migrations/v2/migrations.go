package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
)

func migrateCallbacks(store sdk.KVStore, cdc codec.BinaryCodec) error {
	paramsStore := prefix.NewStore(store, []byte(icacallbacktypes.CallbackDataKeyPrefix))

	iter := paramsStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {

		// Deserialize the callback data
		var oldCallbackData icacallbacktypes.CallbackData
		err := cdc.Unmarshal(iter.Value(), &oldCallbackData)
		if err != nil {
			return err
		}

		// Convert the callback data
		// This will only convert the callback data args, of which the serialization has changed
		newCallbackData, err := convertCallbackData(oldCallbackData)
		if err != nil {
			return err
		}
		newCallbackDataBz, err := cdc.Marshal(&newCallbackData)
		if err != nil {
			return err
		}

		// Set new value on store.
		paramsStore.Set(iter.Key(), newCallbackDataBz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	return migrateCallbacks(store, cdc)
}
