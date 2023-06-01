package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	newData transfertypes.FungibleTokenPacketData,
	packetMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return errorsmod.Wrapf(types.ErrPacketForwardingInactive, "autopilot stakeibc routing is inactive")
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing with STRD tokens
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), newData.Denom) {
		return errors.New("the native token is not supported for liquid staking")
	}

	amount, ok := sdk.NewIntFromString(newData.Amount)
	if !ok {
		return errors.New("not a parsable amount field")
	}

	// Note: newData.denom is base denom e.g. uatom - not ibc/xxx
	var token = sdk.NewCoin(newData.Denom, amount)

	prefixedDenom := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + newData.Denom
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, token.Denom)
	if err != nil {
		return fmt.Errorf("host zone not found for denom (%s)", token.Denom)
	}

	if hostZone.IbcDenom != ibcDenom {
		return fmt.Errorf("ibc denom %s is not equal to host zone ibc denom %s", ibcDenom, hostZone.IbcDenom)
	}

	strideAddress, err := sdk.AccAddressFromBech32(packetMetadata.StrideAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid stride_address (%s) in autopilot memo", strideAddress)
	}

	return k.RunLiquidStake(ctx, strideAddress, token)
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
