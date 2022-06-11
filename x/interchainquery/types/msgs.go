package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// interchainquery message types
const (
	TypeMsgSubmitQueryResponse = "submitqueryresponse"
	TypeMsgQueryBalance        = "querybalance"
)

var (
	_ sdk.Msg = &MsgSubmitQueryResponse{}
	//_ sdk.Msg = &MsgQueryBalance{}
)

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
	fromAddress, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{fromAddress}
}

//----------------------------------------------------------------

//nolint:interfacer
// func NewQueryBalance(chain_id string, address string, denom string, connection_id string, from_address string) *MsgQueryBalance {
// 	return &MsgQueryBalance{ChainId: chain_id, Address: address, Denom: denom, ConnectionId: connection_id, Caller: from_address, Height: 0}
// }

// Route Implements Msg.
// func (msg MsgQueryBalance) Route() string { return RouterKey }

// // Type Implements Msg.
// func (msg MsgQueryBalance) Type() string { return TypeMsgQueryBalance }

// // ValidateBasic Implements Msg.
// func (msg MsgQueryBalance) ValidateBasic() error {
// 	// TODO: check from address

// 	// TODO: check for valid identifier

// 	// TODO: check for valid chain_id

// 	// TODO: check for valid denominations

// 	return nil
// }

// GetSignBytes Implements Msg.
// func (msg MsgQueryBalance) GetSignBytes() []byte {
// 	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
// }

// // GetSigners Implements Msg.
// func (msg MsgQueryBalance) GetSigners() []sdk.AccAddress {
// 	fromAddress, err := sdk.AccAddressFromBech32(msg.Caller)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return []sdk.AccAddress{fromAddress}
// }

//----------------------------------------------------------------
