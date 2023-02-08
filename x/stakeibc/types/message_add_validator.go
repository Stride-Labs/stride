package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgAddValidator = "add_validator"

var _ sdk.Msg = &MsgAddValidator{}

func NewMsgAddValidator(creator string, hostZone string, name string, address string, commission uint64, weight uint64) *MsgAddValidator {
	return &MsgAddValidator{
		Creator:    creator,
		HostZone:   hostZone,
		Name:       name,
		Address:    address,
		Commission: commission,
		Weight:     weight,
	}
}

func (msg *MsgAddValidator) Route() string {
	return RouterKey
}

func (msg *MsgAddValidator) Type() string {
	return TypeMsgAddValidator
}

func (msg *MsgAddValidator) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddValidator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// name validation
	if len(strings.TrimSpace(msg.Name)) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "name is required")
	}
	// commission validation
	if msg.Commission > 100 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "commission must be between 0 and 100")
	}

	return nil
}
