package types

import (
	"github.com/Stride-Labs/stride/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSetNumValidators = "set_num_validators"

var _ sdk.Msg = &MsgSetNumValidators{}

func NewMsgSetNumValidators(creator string, numValidators uint32) *MsgSetNumValidators {
	return &MsgSetNumValidators{
		Creator:       creator,
		NumValidators: numValidators,
	}
}

func (msg *MsgSetNumValidators) Route() string {
	return RouterKey
}

func (msg *MsgSetNumValidators) Type() string {
	return TypeMsgSetNumValidators
}

func (msg *MsgSetNumValidators) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetNumValidators) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetNumValidators) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// num validators validation
	if msg.NumValidators < 1 {
		return sdkerrors.Wrapf(ErrInvalidNumValidator, "number of validators must be greater than 0 (%d provided)", msg.NumValidators)
	}
	return nil
}
