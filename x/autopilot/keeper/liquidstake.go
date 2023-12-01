package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
	packetMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return errorsmod.Wrapf(types.ErrPacketForwardingInactive, "autopilot stakeibc routing is inactive")
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing with STRD tokens
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), fungibleTokenPacketData.Denom) {
		return errors.New("the native token is not supported for liquid staking")
	}

	amount, ok := sdk.NewIntFromString(fungibleTokenPacketData.Amount)
	if !ok {
		return errors.New("not a parsable amount field")
	}

	// Note: fungibleTokenPacketData.denom is base denom e.g. uatom - not ibc/xxx
	var token = sdk.NewCoin(fungibleTokenPacketData.Denom, amount)

	prefixedDenom := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + fungibleTokenPacketData.Denom
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, token.Denom)
	if err != nil {
		return fmt.Errorf("host zone not found for denom (%s)", token.Denom)
	}

	if hostZone.IbcDenom != ibcDenom {
		return fmt.Errorf("ibc denom %s is not equal to host zone ibc denom %s", ibcDenom, hostZone.IbcDenom)
	}

	// fungibleTokenPacketData.Receiver is either the ICS-20 receiver field, or it's parsed from the memo
	// in the receiver field
	// Importantly, this is also the address that receives tokens (fungibleTokenPacketData.Amount above)
	// That means, anyone can send tokens to any address, and cause that address to liquid stake them
	liquidStakingAddress, err := sdk.AccAddressFromBech32(fungibleTokenPacketData.Receiver)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid stride_address (%s) in autopilot memo", liquidStakingAddress)
	}

	return k.RunLiquidStake(ctx, liquidStakingAddress, token)
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
