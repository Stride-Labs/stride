package v2

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldclaimtypes "github.com/Stride-Labs/stride/v5/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v5/x/claim/types"
)

func migrateClaimParams(store sdk.KVStore, cdc codec.Codec) error {
	// Deserialize with old data type
	oldParamsBz := store.Get([]byte(claimtypes.ParamsKey))
	var oldParams oldclaimtypes.Params
	err := cdc.UnmarshalJSON(oldParamsBz, &oldParams)
	if err != nil {
		return fmt.Errorf("unable to unmarshal claims params using old data types: %s", err.Error())
	}

	// Convert and serialize using the new type
	newParams := convertToNewClaimParams(oldParams)
	newParamsBz, err := cdc.MarshalJSON(&newParams)
	if err != nil {
		return fmt.Errorf("unable to marshal claims params using new data types: %s", err.Error())
	}

	// Store new type
	store.Set([]byte(claimtypes.ParamsKey), newParamsBz)
	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.Codec) error {
	store := ctx.KVStore(storeKey)
	return migrateClaimParams(store, cdc)
}
