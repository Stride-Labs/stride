package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRestoreOracleICA = "restore_oracle_ica"

var _ sdk.Msg = &MsgRestoreOracleICA{}

func NewMsgRestoreOracleICA(creator string, oracleMoniker string) *MsgRestoreOracleICA {
	return &MsgRestoreOracleICA{
		Creator:       creator,
		OracleMoniker: oracleMoniker,
	}
}

func (msg *MsgRestoreOracleICA) Route() string {
	return RouterKey
}

func (msg *MsgRestoreOracleICA) Type() string {
	return TypeMsgRestoreOracleICA
}

func (msg *MsgRestoreOracleICA) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRestoreOracleICA) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRestoreOracleICA) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.OracleMoniker == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-moniker is required")
	}
	if strings.Contains(msg.OracleMoniker, " ") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-moniker cannot contain any spaces")
	}

	return nil
}
