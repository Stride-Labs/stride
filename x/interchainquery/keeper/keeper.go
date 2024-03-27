package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/Stride-Labs/stride/v21/utils"
	"github.com/Stride-Labs/stride/v21/x/interchainquery/types"
)

// Keeper of this module maintains collections of registered zones.
type Keeper struct {
	cdc       codec.Codec
	storeKey  storetypes.StoreKey
	callbacks map[string]types.QueryCallbacks
	IBCKeeper *ibckeeper.Keeper
}

// NewKeeper returns a new instance of zones Keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, ibckeeper *ibckeeper.Keeper) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		callbacks: make(map[string]types.QueryCallbacks),
		IBCKeeper: ibckeeper,
	}
}

func (k *Keeper) SetCallbackHandler(module string, handler types.QueryCallbacks) error {
	_, found := k.callbacks[module]
	if found {
		return fmt.Errorf("callback handler already set for %s", module)
	}
	k.callbacks[module] = handler.RegisterICQCallbacks()
	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SubmitICQRequest(ctx sdk.Context, query types.Query, forceUnique bool) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(query.ChainId,
		"Submitting ICQ Request - module=%s, callbackId=%s, connectionId=%s, queryType=%s, timeout_duration=%d",
		query.CallbackModule, query.CallbackId, query.ConnectionId, query.QueryType, query.TimeoutDuration))

	if err := k.ValidateQuery(ctx, query); err != nil {
		return err
	}

	// Set the timeout using the block time and timeout duration
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano() + query.TimeoutDuration.Nanoseconds())
	query.TimeoutTimestamp = timeoutTimestamp

	// Generate and set the query ID - optionally force it to be unique
	query.Id = k.GetQueryId(ctx, query, forceUnique)
	query.RequestSent = false

	// Set the submission height on the Query to the latest light client height
	// In the query response, this will be used to verify that the query wasn't historical
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, query.ConnectionId)
	if !found {
		return errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, query.ConnectionId)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return errorsmod.Wrapf(clienttypes.ErrClientNotFound, connection.ClientId)
	}
	query.SubmissionHeight = clientState.GetLatestHeight().GetRevisionHeight()

	// Save the query to the store
	// If the same query is re-requested, it will get replace in the store with an updated TTL
	//  and the RequestSent bool reset to false
	k.SetQuery(ctx, query)

	return nil
}

// Re-submit an ICQ, generally used after a timeout
func (k *Keeper) RetryICQRequest(ctx sdk.Context, query types.Query) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(query.ChainId,
		"Queuing ICQ Retry - Query Type: %s, Query ID: %s", query.CallbackId, query.Id))

	// Delete old query
	k.DeleteQuery(ctx, query.Id)

	// Submit a new query (with a new ID)
	if err := k.SubmitICQRequest(ctx, query, true); err != nil {
		return errorsmod.Wrapf(err, types.ErrFailedToRetryQuery.Error())
	}

	return nil
}
