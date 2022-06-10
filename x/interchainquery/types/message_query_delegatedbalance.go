package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgQueryDelegatedbalance = "query_delegatedbalance"

var _ sdk.Msg = &MsgQueryDelegatedbalance{}

func NewMsgQueryDelegatedbalance(creator string, chainID string) *MsgQueryDelegatedbalance {
	return &MsgQueryDelegatedbalance{
		Creator: creator,
		ChainID: chainID,
	}
}

func (msg *MsgQueryDelegatedbalance) Route() string {
	return RouterKey
}

func (msg *MsgQueryDelegatedbalance) Type() string {
	return TypeMsgQueryDelegatedbalance
}

func (msg *MsgQueryDelegatedbalance) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgQueryDelegatedbalance) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgQueryDelegatedbalance) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
