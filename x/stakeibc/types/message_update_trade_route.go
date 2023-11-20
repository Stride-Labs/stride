package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateTradeRoute = "update_trade_route"

var _ sdk.Msg = &MsgDeleteTradeRoute{}

func NewMsgUpdateTradeRoute(
	creator string, 
	hostDenom string,
	rewardDenom string,
	poolId uint64, 
	minSwapAmount sdk.Int, 
	maxSwapAmount sdk.Int,
) *MsgUpdateTradeRoute {
	return &MsgUpdateTradeRoute{
		Creator:  creator,
		HostDenom: hostDenom,
		RewardDenom: rewardDenom,
		PoolId: poolId,
		MinSwapAmount: minSwapAmount,
		MaxSwapAmount: maxSwapAmount,
	}
}

func (msg *MsgUpdateTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTradeRoute) Type() string {
	return TypeMsgUpdateTradeRoute
}

func (msg *MsgUpdateTradeRoute) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateTradeRoute) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateTradeRoute) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
