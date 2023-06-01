package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
)

func migrateCallbacks(store sdk.KVStore, cdc codec.BinaryCodec) error {
	icacallbackStore := prefix.NewStore(store, []byte(icacallbacktypes.CallbackDataKeyPrefix))

	iter := icacallbackStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {

		// Deserialize the callback data
		var oldCallbackData icacallbacktypes.CallbackData
		err := cdc.Unmarshal(iter.Value(), &oldCallbackData)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal callback data")
		}

		// Convert the callback data
		// This will only convert the callback data args, of which the serialization has changed
		newCallbackData, err := convertCallbackData(oldCallbackData)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to convert callback data to new schema")
		}
		newCallbackDataBz, err := cdc.Marshal(&newCallbackData)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal callback data")
		}

		// Set new value on store.
		icacallbackStore.Set(iter.Key(), newCallbackDataBz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	return migrateCallbacks(store, cdc)
}
