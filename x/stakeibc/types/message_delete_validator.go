package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgDeleteValidator = "delete_validator"

var _ sdk.Msg = &MsgDeleteValidator{}

func NewMsgDeleteValidator(creator string, hostZone string, valAddr string) *MsgDeleteValidator {
	return &MsgDeleteValidator{
		Creator:  creator,
		HostZone: hostZone,
		ValAddr:  valAddr,
	}
}

func (msg *MsgDeleteValidator) Route() string {
	return RouterKey
}

func (msg *MsgDeleteValidator) Type() string {
	return TypeMsgDeleteValidator
}

func (msg *MsgDeleteValidator) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteValidator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}
