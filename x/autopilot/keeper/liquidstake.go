package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/Stride-Labs/stride/v8/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	newData transfertypes.FungibleTokenPacketData,
	parsedReceiver *types.ParsedReceiver,
	ack ibcexported.Acknowledgement,
) ibcexported.Acknowledgement {
	params := k.GetParams(ctx)
	if !params.Active {
		return channeltypes.NewErrorAcknowledgement(errors.New("packet forwarding param is not active"))
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing with STRD tokens
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), newData.Denom) {
		return channeltypes.NewErrorAcknowledgement(errors.New("the native token is not supported for liquid staking"))
	}

	amount, ok := sdk.NewIntFromString(newData.Amount)
	if !ok {
		return channeltypes.NewErrorAcknowledgement(errors.New("not a parsable amount field"))
	}

	// Note: newData.denom is base denom e.g. uatom - not ibc/xxx
	var token = sdk.NewCoin(newData.Denom, amount)

	prefixedDenom := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + newData.Denom
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, token.Denom)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if hostZone.IbcDenom != ibcDenom {
		return channeltypes.NewErrorAcknowledgement(errors.New("ibc denom is not equal to host zone ibc denom"))
	}

	err = k.RunLiquidStake(ctx, parsedReceiver.StrideAccAddress, token)
	if err != nil {
		ack = channeltypes.NewErrorAcknowledgement(err)
	}
	return ack
}

func (k Keeper) RunLiquidStake(ctx sdk.Context, addr sdk.AccAddress, token sdk.Coin) error {
	msg := &stakeibctypes.MsgLiquidStake{
		Creator:   addr.String(),
		Amount:    token.Amount,
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
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}
	return nil
}
