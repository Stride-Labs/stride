package types

// import (
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
// )

// const TypeMsgQueryExchangerate = "query_exchangerate"

// var _ sdk.Msg = &MsgQueryExchangerate{}

// func NewMsgQueryExchangerate(creator string, chainID string) *MsgQueryExchangerate {
// 	return &MsgQueryExchangerate{
// 		Creator: creator,
// 		ChainId: chainID,
// 	}
// }

// func (msg *MsgQueryExchangerate) Route() string {
// 	return RouterKey
// }

// func (msg *MsgQueryExchangerate) Type() string {
// 	return TypeMsgQueryExchangerate
// }

// func (msg *MsgQueryExchangerate) GetSigners() []sdk.AccAddress {
// 	creator, err := sdk.AccAddressFromBech32(msg.Creator)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return []sdk.AccAddress{creator}
// }

// func (msg *MsgQueryExchangerate) GetSignBytes() []byte {
// 	bz := ModuleCdc.MustMarshalJSON(msg)
// 	return sdk.MustSortJSON(bz)
// }

// func (msg *MsgQueryExchangerate) ValidateBasic() error {
// 	_, err := sdk.AccAddressFromBech32(msg.Creator)
// 	if err != nil {
// 		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
// 	}
// 	return nil
// }
