package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateTradeRoute = "create_trade_route"

var _ sdk.Msg = &MsgCreateTradeRoute{}

func NewMsgCreateTradeRoute(
	creator string,
	hostChainId string,
	hostConnectionId string,
	hostIcaAddress string,
	rewardChainId string,
	rewardConnectionId string,
	rewardIcaAddress string,
	tradeChainId string,
	tradeConnectionId string,
	tradeIcaAddress string,
	hostRewardTransfer string,
	rewardTradeTransfer string,
	tradeHostTransfer string,
	rewardDenomOnHost string,
	rewardDenomOnReward string,
	rewardDenomOnTrade string,
	targetDenomOnTrade string,
	targetDenomOnHost string,
	poolId uint64,
	minSwapAmount sdk.Int,
	maxSwapAmount sdk.Int,
) *MsgCreateTradeRoute {
	return &MsgCreateTradeRoute{
		Creator:  creator,
		HostChainId: hostChainId,
		HostConnectionId: hostConnectionId,
		HostIcaAddress: hostIcaAddress,
		RewardChainId: rewardChainId,
		RewardConnectionId: rewardConnectionId,
		RewardIcaAddress: rewardIcaAddress,
		TradeChainId: tradeChainId,
		TradeConnectionId: tradeConnectionId,
		TradeIcaAddress: tradeIcaAddress,
		HostRewardTransferChannelId: hostRewardTransfer,
		RewardTradeTransferChannelId: rewardTradeTransfer,
		TradeHostTransferChannelId: tradeHostTransfer,
		PoolId: poolId,
		MinSwapAmount: minSwapAmount,
		MaxSwapAmount: maxSwapAmount,
	}
}

func (msg *MsgCreateTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgCreateTradeRoute) Type() string {
	return TypeMsgCreateTradeRoute
}

func (msg *MsgCreateTradeRoute) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateTradeRoute) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateTradeRoute) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// need to add a lot of validation logic here to verify all trade route fields valid

	return nil
}
