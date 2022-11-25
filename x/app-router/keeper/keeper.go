package keeper

import (
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v3/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v3/x/stakeibc/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v3/x/app-router/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		Cdc            codec.BinaryCodec
		storeKey       sdk.StoreKey
		paramstore     paramtypes.Subspace
		stakeibcKeeper stakeibckeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	ps paramtypes.Subspace,
	stakeibcKeeper stakeibckeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Cdc:            Cdc,
		storeKey:       storeKey,
		paramstore:     ps,
		stakeibcKeeper: stakeibcKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// TODO: Add a LiquidStake function call (maybe)
func (k Keeper) LiquidStakeTransferPacket(ctx sdk.Context, parsedReceiver *types.ParsedReceiver, token sdk.Coin, labels []metrics.Label) error {
	msg := &stakeibctypes.MsgLiquidStake{
		// TODO: do we need a creator here?
		// we could use the recipient...
		// it's a bit strange because this address didn't "create" the liquid stake transaction
		// TODO: check that we don't have assumptions around the creator of a message
		Creator:   parsedReceiver.StrideAccAddress.String(),
		Amount:    token.Amount.Uint64(),
		HostDenom: token.Denom,
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	_, err := msgServer.LiquidStake(
		// goCtx
		sdk.WrapSDKContext(ctx),
		// MsgLiquidStake
		msg,
	)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	defer func() {
		telemetry.SetGaugeWithLabels(
			[]string{"tx", "msg", "ibc", "transfer"},
			float32(token.Amount.Int64()),
			[]metrics.Label{telemetry.NewLabel("label_denom", token.Denom)},
		)

		telemetry.IncrCounterWithLabels(
			[]string{"ibc", types.ModuleName, "send"},
			1,
			labels,
		)
	}()
	return nil
}
