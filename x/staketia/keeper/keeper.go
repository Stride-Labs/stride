package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/staketia/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	icaOracleKeeper types.ICAOracleKeeper
	ratelimitKeeper types.RatelimitKeeper
	recordsKeeper   types.RecordsKeeper
	stakeibcKeeper  types.StakeibcKeeper
	transferKeeper  types.TransferKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	icaOracleKeeper types.ICAOracleKeeper,
	ratelimitKeeper types.RatelimitKeeper,
	recordsKeeper types.RecordsKeeper,
	stakeibcKeeper types.StakeibcKeeper,
	transferKeeper types.TransferKeeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		icaOracleKeeper: icaOracleKeeper,
		ratelimitKeeper: ratelimitKeeper,
		recordsKeeper:   recordsKeeper,
		stakeibcKeeper:  stakeibcKeeper,
		transferKeeper:  transferKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
