package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgAddTokenPrice    = "add_token_price"
	TypeMsgRemoveTokenPrice = "remove_token_price"
)

var (
	_ sdk.Msg = &MsgAddTokenPrice{}
	_ sdk.Msg = &MsgRemoveTokenPrice{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgAddTokenPrice{}
	_ legacytx.LegacyMsg = &MsgRemoveTokenPrice{}
)

// ----------------------------------------------
//               MsgClaim
// ----------------------------------------------

func NewMsgAddTokenPrice(sender, baseDenom, quoteDenom string) *MsgAddTokenPrice {
	return &MsgAddTokenPrice{
		Sender:     sender,
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
}

func (msg MsgAddTokenPrice) Type() string {
	return TypeMsgAddTokenPrice
}

func (msg MsgAddTokenPrice) Route() string {
	return RouterKey
}

func (msg *MsgAddTokenPrice) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgAddTokenPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddTokenPrice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.BaseDenom == "" {
		return errors.New("base-denom must be specified")
	}
	if msg.QuoteDenom == "" {
		return errors.New("quote-denom must be specified")
	}

	return nil
}

// ----------------------------------------------
//               MsgRemoveTokenPrice
// ----------------------------------------------

func NewMsgRemoveTokenPrice(sender, baseDenom, quoteDenom string) *MsgRemoveTokenPrice {
	return &MsgRemoveTokenPrice{
		Sender:     sender,
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
}

func (msg MsgRemoveTokenPrice) Type() string {
	return TypeMsgRemoveTokenPrice
}

func (msg MsgRemoveTokenPrice) Route() string {
	return RouterKey
}

func (msg *MsgRemoveTokenPrice) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgRemoveTokenPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveTokenPrice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.BaseDenom == "" {
		return errors.New("base-denom must be specified")
	}
	if msg.QuoteDenom == "" {
		return errors.New("quote-denom must be specified")
	}

	return nil
}
