package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRestoreOracleICA = "restore_oracle_ica"

var _ sdk.Msg = &MsgRestoreOracleICA{}

func NewMsgRestoreOracleICA(creator string, oracleChainId string) *MsgRestoreOracleICA {
	return &MsgRestoreOracleICA{
		Creator:       creator,
		OracleChainId: oracleChainId,
	}
}

func (msg MsgRestoreOracleICA) Type() string {
	return TypeMsgRestoreOracleICA
}

func (msg MsgRestoreOracleICA) Route() string {
	return RouterKey
}

func (msg *MsgRestoreOracleICA) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRestoreOracleICA) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.OracleChainId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-chain-id is required")
	}

	return nil
}
