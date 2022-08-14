package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	icqkeeper "github.com/Stride-Labs/stride/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	icacallbacksmodulekeeper "github.com/Stride-Labs/stride/x/icacallbacks/keeper"
	recordsmodulekeeper "github.com/Stride-Labs/stride/x/records/keeper"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc                   codec.BinaryCodec
		storeKey              sdk.StoreKey
		memKey                sdk.StoreKey
		paramstore            paramtypes.Subspace
		ICAControllerKeeper   icacontrollerkeeper.Keeper
		IBCKeeper             ibckeeper.Keeper
		scopedKeeper          capabilitykeeper.ScopedKeeper
		TransferKeeper        ibctransferkeeper.Keeper
		bankKeeper            bankkeeper.Keeper
		InterchainQueryKeeper icqkeeper.Keeper
		RecordsKeeper         recordsmodulekeeper.Keeper
		StakingKeeper         stakingkeeper.Keeper
		ICACallbacksKeeper    icacallbacksmodulekeeper.Keeper

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
	RecordsKeeper recordsmodulekeeper.Keeper,
	StakingKeeper stakingkeeper.Keeper,
	ICACallbacksKeeper icacallbacksmodulekeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                   cdc,
		storeKey:              storeKey,
		memKey:                memKey,
		paramstore:            ps,
		accountKeeper:         accountKeeper,
		bankKeeper:            bankKeeper,
		ICAControllerKeeper:   icacontrollerkeeper,
		IBCKeeper:             ibcKeeper,
		scopedKeeper:          scopedKeeper,
		TransferKeeper:        TransferKeeper,
		InterchainQueryKeeper: interchainQueryKeeper,
		RecordsKeeper:         RecordsKeeper,
		StakingKeeper:         StakingKeeper,
		ICACallbacksKeeper:    ICACallbacksKeeper,
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

// helper to get what share of the curr epoch we're through
func (k Keeper) GetStrideEpochElapsedShare(ctx sdk.Context) (sdk.Dec, error) {
	epochType := epochstypes.STRIDE_EPOCH
	epochTracker, found := k.GetEpochTracker(ctx, epochType)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to get epoch tracker for %s", epochType))
		return sdk.ZeroDec(), sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to get epoch tracker for %s", epochType)
	}
	prevEpochStartTime := epochTracker.NextEpochStartTime - epochTracker.Duration
	elapsedShare := sdk.NewDec(ctx.BlockTime().UnixNano() - prevEpochStartTime).Quo(sdk.NewDec(epochTracker.Duration))
	k.Logger(ctx).Info(fmt.Sprintf("epochTracker.NextEpochStartTime %v epochTracker.Duration %v prevEpochStartTime %v", epochTracker.NextEpochStartTime, epochTracker.Duration, prevEpochStartTime))
	return elapsedShare, nil
}

// helper to check whether ICQs are valid in this portion of the epoch
func (k Keeper) IsWithinBufferWindow(ctx sdk.Context) (bool, error) {
	elapsedShareOfEpoch, err := k.GetStrideEpochElapsedShare(ctx)
	if err != nil {
		return false, err
	}
	bufferSize, err := cast.ToInt64E(k.GetParam(ctx, types.KeyBufferSize))
	if err != nil {
		return false, err
	}
	epochShareThresh := sdk.NewDec(1).Sub(sdk.NewDec(1).Quo(sdk.NewDec(bufferSize)))

	inWindow := elapsedShareOfEpoch.GT(epochShareThresh)
	if !inWindow {
		k.Logger(ctx).Error(fmt.Sprintf("ICQCB: We're %d pct through the epoch, ICQ cutoff is %d", elapsedShareOfEpoch, epochShareThresh))
	}
	return inWindow, nil
}
