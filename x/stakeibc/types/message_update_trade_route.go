package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateTradeRoute = "update_trade_route"

var _ sdk.Msg = &MsgUpdateTradeRoute{}

func (msg *MsgUpdateTradeRoute) Type() string {
	return TypeMsgUpdateTradeRoute
}

func (msg *MsgUpdateTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTradeRoute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateTradeRoute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if msg.HostDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "missing host denom")
	}
	if msg.RewardDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "missing reward denom")
	}

	if msg.MinTransferAmount.IsNil() || msg.MinTransferAmount.LT(sdkmath.ZeroInt()) {
		return errors.New("min transfer amount must be greater than or equal to zero")
	}

	return nil
}
