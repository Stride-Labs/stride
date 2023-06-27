package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v11/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v11/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

func (k Keeper) TryRedeemStake(
	ctx sdk.Context,
	packet channeltypes.Packet,
	newData transfertypes.FungibleTokenPacketData,
	packetMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return fmt.Errorf("packet forwarding param is not active")
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing IBC tokens from other chains
	if !transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), newData.Denom) {
		return fmt.Errorf("the ibc tokens are not supported for redeem stake")
	}

	voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	stAssetDenom := newData.Denom[len(voucherPrefix):]
	if !stakeibctypes.IsStAssetDenom(stAssetDenom) {
		return fmt.Errorf("not a liquid staking token")
	}

	hostZoneDenom := stakeibctypes.HostZoneDenomFromStAssetDenom(stAssetDenom)

	amount, ok := sdk.NewIntFromString(newData.Amount)
	if !ok {
		return fmt.Errorf("not a parsable amount field")
	}

	// Note: newData.denom is ibc denom for st assets - e.g. ibc/xxx
	var token = sdk.Coin{
		Denom:  newData.Denom,
		Amount: amount,
	}

	if err := token.Validate(); err != nil {
		return err
	}

	strideAddress, err := sdk.AccAddressFromBech32(newData.Receiver)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver (%s) in autopilot memo", strideAddress)
	}

	return k.RunRedeemStake(ctx, strideAddress, packetMetadata.IbcReceiver, hostZoneDenom, token, []metrics.Label{})
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
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}
	return nil
}
