package keeper

import (
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/v3/x/app-router/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v3/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	newData transfertypes.FungibleTokenPacketData,
	parsedReceiver *types.ParsedReceiver,
	ack ibcexported.Acknowledgement,
) ibcexported.Acknowledgement {
	// recalculate denom, skip checks that were already done in app.OnRecvPacket
	var err error
	// TODO put denom handling in separate function
	var denom string
	// in this case, we can't process a liquid staking transaction, because we're dealing with STRD tokens
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), newData.Denom) {
		// remove prefix added by sender chain
		voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := newData.Denom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom

		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}
		// TODO: can we just delete the above code?
		return ack
	} else {
		prefixedDenom := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + newData.Denom
		denom = transfertypes.ParseDenomTrace(prefixedDenom).BaseDenom
	}
	amount, ok := sdk.NewIntFromString(newData.Amount)
	if !ok {
		channeltypes.NewErrorAcknowledgement("not a parsable amount field")
	}
	var token = sdk.NewCoin(denom, amount)

	err = k.RunLiquidStake(ctx, parsedReceiver.StrideAccAddress, token, []metrics.Label{})
	if err != nil {
		ack = channeltypes.NewErrorAcknowledgement(err.Error())
	}
	return ack
}

func (k Keeper) RunLiquidStake(ctx sdk.Context, addr sdk.AccAddress, token sdk.Coin, labels []metrics.Label) error {
	msg := &stakeibctypes.MsgLiquidStake{
		// TODO: do we need a creator here?
		// we could use the recipient...
		// it's a bit strange because this address didn't "create" the liquid stake transaction
		// TODO: check that we don't have assumptions around the creator of a message
		Creator:   addr.String(),
		Amount:    token.Amount.Uint64(),
		HostDenom: token.Denom,
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	_, err := msgServer.LiquidStake(
		sdk.WrapSDKContext(ctx),
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
