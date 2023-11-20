package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeleteTradeRoute = "delete_trade_route"

var _ sdk.Msg = &MsgDeleteTradeRoute{}

func NewMsgDeleteTradeRoute(creator string, hostDenom string, rewardDenom string) *MsgDeleteTradeRoute {
	return &MsgDeleteTradeRoute{
		Creator:  creator,
		HostDenom: hostDenom,
		RewardDenom: rewardDenom,
	}
}

func (msg *MsgDeleteTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgDeleteTradeRoute) Type() string {
	return TypeMsgDeleteTradeRoute
}

func (msg *MsgDeleteTradeRoute) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteTradeRoute) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteTradeRoute) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.HostDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "missing host denom")
	}
	if msg.RewardDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "missing reward denom")
	}

	return nil
}
