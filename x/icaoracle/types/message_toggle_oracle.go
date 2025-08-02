package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgToggleOracle = "toggle_oracle"

var _ sdk.Msg = &MsgToggleOracle{}

func (msg MsgToggleOracle) Type() string {
	return TypeMsgToggleOracle
}

func (msg MsgToggleOracle) Route() string {
	return RouterKey
}

func (msg *MsgToggleOracle) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgToggleOracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if msg.OracleChainId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-chain-id is required")
	}

	return nil
}
