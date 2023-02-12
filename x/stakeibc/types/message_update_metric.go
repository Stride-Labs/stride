package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateMetric = "update_metrics"

var _ sdk.Msg = &MsgUpdateMetric{}

func NewMsgUpdateMetric(creator string, contractAddress string) *MsgUpdateMetric {
	return &MsgUpdateMetric{
		Creator:         creator,
		ContractAddress: contractAddress,
	}
}

func (msg *MsgUpdateMetric) Route() string {
	return RouterKey
}

func (msg *MsgUpdateMetric) Type() string {
	return TypeMsgUpdateMetric
}

func (msg *MsgUpdateMetric) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateMetric) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateMetric) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
