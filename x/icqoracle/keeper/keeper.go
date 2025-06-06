package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/icqoracle/types"
)

type Keeper struct {
	cdc               codec.Codec
	storeKey          storetypes.StoreKey
	IcqKeeper         types.IcqKeeper
	ibcTransferKeeper types.IbcTransferKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	icqKeeper types.IcqKeeper,
	ibcTransferKeeper types.IbcTransferKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		IcqKeeper:         icqKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
		authority:         authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the x/icqoracle module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
