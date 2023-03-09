package keeper

import (
	"fmt"

	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/Stride-Labs/stride/v6/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v6/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

func (k Keeper) TryRedeemStake(
	ctx sdk.Context,
	packet channeltypes.Packet,
	newData transfertypes.FungibleTokenPacketData,
	parsedReceiver *types.ParsedReceiver,
	ack ibcexported.Acknowledgement,
) ibcexported.Acknowledgement {
	fmt.Println("Autopilot.TryRedeemStake1")
	params := k.GetParams(ctx)
	if !params.Active {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("packet forwarding param is not active"))
	}

	fmt.Println("Autopilot.TryRedeemStake2")
	// In this case, we can't process a liquid staking transaction, because we're dealing IBC tokens from other chains
	if !transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), newData.Denom) {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("the ibc tokens are not supported for redeem stake"))
	}

	voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	stAssetDenom := newData.Denom[len(voucherPrefix):]
	if !stakeibctypes.IsStAssetDenom(stAssetDenom) {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("not a liquid staking token"))
	}

	fmt.Println("Autopilot.TryRedeemStake3")
	hostZoneDenom := stakeibctypes.HostZoneDenomFromStAssetDenom(stAssetDenom)

	amount, ok := sdk.NewIntFromString(newData.Amount)
	if !ok {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("not a parsable amount field"))
	}

	fmt.Println("Autopilot.TryRedeemStake4")
	// Note: newData.denom is ibc denom for st assets - e.g. ibc/xxx
	var token = sdk.NewCoin(newData.Denom, amount)

	err := k.RunRedeemStake(ctx, parsedReceiver.StrideAccAddress, parsedReceiver.ResultReceiver, hostZoneDenom, token, []metrics.Label{})
	if err != nil {
		ack = channeltypes.NewErrorAcknowledgement(err)
	}
	fmt.Println("Autopilot.TryRedeemStake5")
	return ack
}

func (k Keeper) RunRedeemStake(ctx sdk.Context, addr sdk.AccAddress, receiver string, hostZoneDenom string, token sdk.Coin, labels []metrics.Label) error {
	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, hostZoneDenom)
	if err != nil {
		return err
	}

	msg := &stakeibctypes.MsgRedeemStake{
		Creator:  addr.String(),
		Amount:   token.Amount,
		HostZone: hostZone.ChainId,
		Receiver: receiver,
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	_, err = msgServer.RedeemStake(
		sdk.WrapSDKContext(ctx),
		msg,
	)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}
	return nil
}
