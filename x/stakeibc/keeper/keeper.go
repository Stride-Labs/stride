package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	icqkeeper "github.com/Stride-Labs/stride/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc                 codec.BinaryCodec
		storeKey            sdk.StoreKey
		memKey              sdk.StoreKey
		paramstore          paramtypes.Subspace
		ICAControllerKeeper icacontrollerkeeper.Keeper
		IBCKeeper           ibckeeper.Keeper
		scopedKeeper        capabilitykeeper.ScopedKeeper
		TransferKeeper      ibctransferkeeper.Keeper
		bankKeeper    		bankkeeper.Keeper
		InterchainQueryKeeper	icqkeeper.Keeper

		accountKeeper types.AccountKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	// channelKeeper cosmosibckeeper.ChannelKeeper,
	// portKeeper cosmosibckeeper.PortKeeper,
	// scopedKeeper cosmosibckeeper.ScopedKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	icacontrollerkeeper icacontrollerkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	TransferKeeper ibctransferkeeper.Keeper,
	interchainQueryKeeper icqkeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		paramstore:          ps,
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		ICAControllerKeeper: icacontrollerkeeper,
		IBCKeeper:           ibcKeeper,
		scopedKeeper:        scopedKeeper,
		TransferKeeper:      TransferKeeper,
		InterchainQueryKeeper: interchainQueryKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) GetChainID(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return "", fmt.Errorf("invalid connection id, \"%s\" not found", connectionID)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		return "", fmt.Errorf("client id \"%s\" not found for connection \"%s\"", conn.ClientId, connectionID)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", fmt.Errorf("invalid client state for client \"%s\" on connection \"%s\"", conn.ClientId, connectionID)
	}

	return client.ChainId, nil
}

func (k Keeper) GetCounterpartyChainId(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return "", fmt.Errorf("invalid connection id, \"%s\" not found", connectionID)
	}
	counterPartyClientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.Counterparty.ClientId)
	if !found {
		return "", fmt.Errorf("counterparty client id \"%s\" not found for connection \"%s\"", conn.Counterparty.ClientId, connectionID)
	}
	counterpartyClient, ok := counterPartyClientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", fmt.Errorf("invalid client state for client \"%s\" on connection \"%s\"", conn.Counterparty.ClientId, connectionID)
	}

	return counterpartyClient.ChainId, nil
}

func (k Keeper) GetConnectionId(ctx sdk.Context, portId string) (string, error) {
	icas := k.ICAControllerKeeper.GetAllInterchainAccounts(ctx)
	for _, ica := range icas {
		if ica.PortId == portId {
			return ica.ConnectionId, nil
		}
	}
	return "", fmt.Errorf("portId %s has no associated connectionId", portId)
}
