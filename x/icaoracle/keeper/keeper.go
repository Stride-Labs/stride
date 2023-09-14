package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramstore paramtypes.Subspace
	authority  string

	ICS4Wrapper         types.ICS4Wrapper
	ClientKeeper        types.ClientKeeper
	ConnectionKeeper    types.ConnectionKeeper
	ChannelKeeper       types.ChannelKeeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICACallbacksKeeper  types.ICACallbacksKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramstore paramtypes.Subspace,
	authority string,

	ics4Wrapper types.ICS4Wrapper,
	clientKeeper types.ClientKeeper,
	connectionKeeper types.ConnectionKeeper,
	channelKeeper types.ChannelKeeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	icaCallbacksKeeper types.ICACallbacksKeeper,
) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   key,
		paramstore: paramstore,
		authority:  authority,

		ICS4Wrapper:         ics4Wrapper,
		ClientKeeper:        clientKeeper,
		ConnectionKeeper:    connectionKeeper,
		ChannelKeeper:       channelKeeper,
		ICAControllerKeeper: icaControllerKeeper,
		ICACallbacksKeeper:  icaCallbacksKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the x/staking module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
