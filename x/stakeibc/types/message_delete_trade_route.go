package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeleteTradeRoute = "delete_trade_route"

var _ sdk.Msg = &MsgDeleteTradeRoute{}

func (msg *MsgDeleteTradeRoute) Type() string {
	return TypeMsgDeleteTradeRoute
}

func (msg *MsgDeleteTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgDeleteTradeRoute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgDeleteTradeRoute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if msg.HostDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "missing host denom")
	}
	if msg.RewardDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "missing reward denom")
	}

	return nil
}
