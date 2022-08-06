package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// interchainquery message types
const (
	TypeMsgSubmitQueryResponse = "submitqueryresponse"
)

var (
	_ sdk.Msg = &MsgSubmitQueryResponse{}
)

// NewMsgSubmitQueryResponse - construct a msg to fulfil query request.
//nolint:interfacer
func NewMsgSubmitQueryResponse(chain_id string, result string, from_address sdk.Address) *MsgSubmitQueryResponse {
	return &MsgSubmitQueryResponse{ChainId: chain_id, Result: nil, FromAddress: from_address.String()}
}

// Route Implements Msg.
func (msg MsgSubmitQueryResponse) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSubmitQueryResponse) Type() string { return TypeMsgSubmitQueryResponse }

// ValidateBasic Implements Msg.
func (msg MsgSubmitQueryResponse) ValidateBasic() error {
	// check from address
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid fromAddress in ICQ response (%s)", err)
	}
	// check chain_id is not empty
	if msg.ChainId == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "chain_id cannot be empty in ICQ response")
	}
	// check amound positiv
	if !m.Coins.IsAllPositive() {
		return fmt.Errorf("cannot lock up a zero or negative amount")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSubmitQueryResponse) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgSubmitQueryResponse) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}
