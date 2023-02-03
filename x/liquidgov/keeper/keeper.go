package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v5/modules/core/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"

	icacallbackskeeper "github.com/Stride-Labs/stride/v5/x/icacallbacks/keeper"
	icqkeeper "github.com/Stride-Labs/stride/v5/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		stakeibcKeeper        types.StakeibcKeeper
		accountKeeper         types.AccountKeeper
		bankKeeper            types.BankKeeper
		InterchainQueryKeeper icqkeeper.Keeper
		ICACallbacksKeeper    icacallbackskeeper.Keeper
		ICAControllerKeeper   icacontrollerkeeper.Keeper
		scopedKeeper          capabilitykeeper.ScopedKeeper
		IBCKeeper             ibckeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	stakeibcKeeper types.StakeibcKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	interchainQueryKeeper icqkeeper.Keeper,
	ICACallbacksKeeper icacallbackskeeper.Keeper,
	icacontrollerkeeper icacontrollerkeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	ibcKeeper ibckeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		stakeibcKeeper:        stakeibcKeeper,
		accountKeeper:         accountKeeper,
		bankKeeper:            bankKeeper,
		InterchainQueryKeeper: interchainQueryKeeper,
		ICACallbacksKeeper:    ICACallbacksKeeper,
		ICAControllerKeeper:   icacontrollerkeeper,
		scopedKeeper:          scopedKeeper,
		IBCKeeper:             ibcKeeper,
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
		errMsg := fmt.Sprintf("invalid connection id, %s not found", connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		errMsg := fmt.Sprintf("client id %s not found for connection %s", conn.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		errMsg := fmt.Sprintf("invalid client state for client %s on connection %s", conn.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	return client.ChainId, nil
}

func (k Keeper) GetCounterpartyChainId(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		errMsg := fmt.Sprintf("invalid connection id, %s not found", connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	counterPartyClientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.Counterparty.ClientId)
	if !found {
		errMsg := fmt.Sprintf("counterparty client id %s not found for connection %s", conn.Counterparty.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	counterpartyClient, ok := counterPartyClientState.(*ibctmtypes.ClientState)
	if !ok {
		errMsg := fmt.Sprintf("invalid client state for client %s on connection %s", conn.Counterparty.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
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
	errMsg := fmt.Sprintf("portId %s has no associated connectionId", portId)
	k.Logger(ctx).Error(errMsg)
	return "", fmt.Errorf(errMsg)
}
