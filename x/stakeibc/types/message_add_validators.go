package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v9/utils"
)

const TypeMsgAddValidators = "add_validators"

var _ sdk.Msg = &MsgAddValidators{}

func NewMsgAddValidators(creator string, hostZone string, validators []*Validator) *MsgAddValidators {
	return &MsgAddValidators{
		Creator:    creator,
		HostZone:   hostZone,
		Validators: validators,
	}
}

func (msg *MsgAddValidators) Route() string {
	return RouterKey
}

func (msg *MsgAddValidators) Type() string {
	return TypeMsgAddValidators
}

func (msg *MsgAddValidators) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddValidators) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddValidators) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	if len(msg.Validators) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least one validator must be provided")
	}

	for i, validator := range msg.Validators {
		if len(strings.TrimSpace(validator.Name)) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator name is required (index %d)", i)
		}
		if len(strings.TrimSpace(validator.Address)) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address is required (index %d)", i)
		}
	}

	return nil
}
