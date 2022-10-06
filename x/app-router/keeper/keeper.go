package keeper

import (
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"

	"github.com/Stride-Labs/stride/x/app_router/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		Cdc            codec.BinaryCodec
		storeKey       sdk.StoreKey
		memKey         sdk.StoreKey
		paramstore     paramtypes.Subspace
		scopedKeeper   capabilitykeeper.ScopedKeeper
		stakeibcKeeper stakeibckeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Cdc:            Cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		paramstore:     ps,
		scopedKeeper:   scopedKeeper,
		stakeibcKeeper: stakeibcKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

// TODO: Add a LiquidStake function call (maybe)
func (k Keeper) LiquidStakeTransferPacket(ctx sdk.Context, parsedReceiver *types.ParsedReceiver, token sdk.Coin, labels []metrics.Label) error {
	msg := stakeibctypes.MsgLiquidStake{
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
	err := k.stakeibcKeeper.LiquidStake(
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
			[]metrics.Label{telemetry.NewLabel(coretypes.LabelDenom, token.Denom)},
		)

		telemetry.IncrCounterWithLabels(
			[]string{"ibc", types.ModuleName, "send"},
			1,
			labels,
		)
	}()
	return nil
}
