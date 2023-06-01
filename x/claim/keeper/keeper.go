package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/claim/types"
)

// Keeper struct
type Keeper struct {
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	distrKeeper   types.DistrKeeper
	epochsKeeper  types.EpochsKeeper
}

// NewKeeper returns keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, dk types.DistrKeeper, ek types.EpochsKeeper) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
		distrKeeper:   dk,
		epochsKeeper:  ek,
	}
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
