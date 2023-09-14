package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// Stores/updates an oracle object in the store
func (k Keeper) SetOracle(ctx sdk.Context, oracle types.Oracle) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)

	oracleKey := types.KeyPrefix(oracle.ChainId)
	oracleValue := k.cdc.MustMarshal(&oracle)

	store.Set(oracleKey, oracleValue)
}

// Grabs and returns an oracle object from the store using the chain-id
func (k Keeper) GetOracle(ctx sdk.Context, chainId string) (oracle types.Oracle, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)

	oracleKey := types.KeyPrefix(chainId)
	oracleBz := store.Get(oracleKey)

	if len(oracleBz) == 0 {
		return oracle, false
	}

	k.cdc.MustUnmarshal(oracleBz, &oracle)
	return oracle, true
}

// Returns all oracles
func (k Keeper) GetAllOracles(ctx sdk.Context) []types.Oracle {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allOracles := []types.Oracle{}
	for ; iterator.Valid(); iterator.Next() {

		oracle := types.Oracle{}
		k.cdc.MustUnmarshal(iterator.Value(), &oracle)
		allOracles = append(allOracles, oracle)
	}

	return allOracles
}

// Removes an oracle from the store
func (k Keeper) RemoveOracle(ctx sdk.Context, chainId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)
	oracleKey := types.KeyPrefix(chainId)
	store.Delete(oracleKey)
}

// Toggle whether an oracle is active
func (k Keeper) ToggleOracle(ctx sdk.Context, chainId string, active bool) error {
	oracle, found := k.GetOracle(ctx, chainId)
	if !found {
		return types.ErrOracleNotFound
	}

	// If the oracle is being set to active, we need to first validate the ICA setup
	if active {
		if err := oracle.ValidateICASetup(); err != nil {
			return err
		}
		if err := oracle.ValidateContractInstantiated(); err != nil {
			return err
		}
		if !k.IsOracleICAChannelOpen(ctx, oracle) {
			return errorsmod.Wrapf(types.ErrOracleICAChannelClosed,
				"chain-id: %s, channel-id: %s", chainId, oracle.ChannelId)
		}
	}

	oracle.Active = active
	k.SetOracle(ctx, oracle)
	return nil
}

// Grab's an oracle from it's connectionId
func (k Keeper) GetOracleFromConnectionId(ctx sdk.Context, connectionId string) (oracle types.Oracle, found bool) {
	for _, oracle := range k.GetAllOracles(ctx) {
		if oracle.ConnectionId == connectionId {
			return oracle, true
		}
	}
	return oracle, false
}

// Checks if the oracle ICA channel is open
func (k Keeper) IsOracleICAChannelOpen(ctx sdk.Context, oracle types.Oracle) bool {
	channel, found := k.ChannelKeeper.GetChannel(ctx, oracle.PortId, oracle.ChannelId)
	if !found {
		return false
	}
	return channel.State == channeltypes.OPEN
}
