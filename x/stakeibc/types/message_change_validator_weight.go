package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgChangeValidatorWeight = "change_validator_weight"

var _ sdk.Msg = &MsgChangeValidatorWeight{}

func NewMsgChangeValidatorWeight(creator string, hostZone string, address string, weight uint64) *MsgChangeValidatorWeight {
	return &MsgChangeValidatorWeight{
		Creator:  creator,
		HostZone: hostZone,
		ValAddr:  address,
		Weight:   weight,
	}
}

func (msg *MsgChangeValidatorWeight) Route() string {
	return RouterKey
}

func (msg *MsgChangeValidatorWeight) Type() string {
	return TypeMsgChangeValidatorWeight
}

func (msg *MsgChangeValidatorWeight) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgChangeValidatorWeight) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgChangeValidatorWeight) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
