package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// interchainquery message types
const (
	TypeMsgSubmitQueryResponse = "submitqueryresponse"
)

var _ sdk.Msg = &MsgSubmitQueryResponse{}

// NewMsgSubmitQueryResponse - construct a msg to fulfil query request.
//nolint:interfacer
func NewMsgSubmitQueryResponse(chain_id string, result string, from_address sdk.Address) *MsgSubmitQueryResponse {
	// TODO: fix me.
	return &MsgSubmitQueryResponse{ChainId: chain_id, Result: nil, FromAddress: from_address.String()}
}

// Route Implements Msg.
func (msg MsgSubmitQueryResponse) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSubmitQueryResponse) Type() string { return TypeMsgSubmitQueryResponse }

// ValidateBasic Implements Msg.
func (msg MsgSubmitQueryResponse) ValidateBasic() error {
	// TODO: check from address

	// TODO: check for valid identifier

	// TODO: check for valid chain_id

	// TODO: check for valid denominations

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
