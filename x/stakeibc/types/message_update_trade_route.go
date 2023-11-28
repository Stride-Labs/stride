package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateTradeRoute = "update_trade_route"

var (
	_ sdk.Msg            = &MsgUpdateTradeRoute{}
	_ legacytx.LegacyMsg = &MsgUpdateTradeRoute{}
)

func (msg *MsgUpdateTradeRoute) Type() string {
	return TypeMsgUpdateTradeRoute
}

func (msg *MsgUpdateTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTradeRoute) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
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

	if msg.PoolId < 1 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	if msg.MinSwapAmount > msg.MaxSwapAmount {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "min swap amount cannot be greater than max swap amount")
	}

	maxAllowedSwapLossRate, err := sdk.NewDecFromStr(msg.MaxAllowedSwapLossRate)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to cast max allowed swap loss rate to a decimal")
	}
	if maxAllowedSwapLossRate.LT(sdk.ZeroDec()) || maxAllowedSwapLossRate.GT(sdk.OneDec()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max allowed swap loss rate must be between 0 and 1")
	}

	return nil
}
