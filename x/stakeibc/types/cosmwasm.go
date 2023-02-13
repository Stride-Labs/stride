package types

import (
	"bytes"
	"encoding/json"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// MaxLabelSize is the longest label that can be used when instantiating a contract
	MaxLabelSize = 128 // extension point for chains to customize via compile flag.
)

// ValidateLabel ensure label constraints
func ValidateLabel(label string) error {
	if label == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "is required")
	}
	if len(label) > MaxLabelSize {
		return sdkerrors.ErrInvalidRequest.Wrapf("cannot be longer than %d characters", MaxLabelSize)
	}
	return nil
}

// RawContractMessage defines a json message that is sent or returned by a wasm contract.
// This type can hold any type of bytes. Until validateBasic is called there should not be
// any assumptions made that the data is valid syntax or semantic.
type RawContractMessage []byte

func (r RawContractMessage) MarshalJSON() ([]byte, error) {
	return json.RawMessage(r).MarshalJSON()
}

func (r *RawContractMessage) UnmarshalJSON(b []byte) error {
	if r == nil {
		return errors.New("unmarshalJSON on nil pointer")
	}
	*r = append((*r)[0:0], b...)
	return nil
}

func (r *RawContractMessage) ValidateBasic() error {
	if r == nil {
		return sdkerrors.ErrInvalidType
	}
	if !json.Valid(*r) {
		return sdkerrors.ErrInvalidRequest
	}
	return nil
}

// Bytes returns raw bytes type
func (r RawContractMessage) Bytes() []byte {
	return r
}

// Equal content is equal json. Byte equal but this can change in the future.
func (r RawContractMessage) Equal(o RawContractMessage) bool {
	return bytes.Equal(r.Bytes(), o.Bytes())
}

var _ sdk.Msg = &MsgExecuteContract{}

func (msg MsgExecuteContract) Route() string {
	return RouterKey
}

func (msg MsgExecuteContract) Type() string {
	return "execute"
}

func (msg MsgExecuteContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(err, "sender")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Contract); err != nil {
		return sdkerrors.Wrap(err, "contract")
	}

	if !msg.Funds.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "sentFunds")
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "payload msg")
	}
	return nil
}

func (msg MsgExecuteContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgExecuteContract) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgInstantiateContract) Route() string {
	return RouterKey
}

func (msg MsgInstantiateContract) Type() string {
	return "instantiate"
}

func (msg MsgInstantiateContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(err, "sender")
	}

	if msg.CodeID == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "code id is required")
	}

	if err := ValidateLabel(msg.Label); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "label is required")
	}

	if !msg.Funds.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}

	if len(msg.Admin) != 0 {
		if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
			return sdkerrors.Wrap(err, "admin")
		}
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "payload msg")
	}
	return nil
}

func (msg MsgInstantiateContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgInstantiateContract) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
