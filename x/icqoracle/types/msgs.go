package types

import (
	"errors"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgRegisterTokenPriceQuery = "register_token_price_query"
	TypeMsgRemoveTokenPrice        = "remove_token_price"
)

var (
	_ sdk.Msg = &MsgRegisterTokenPriceQuery{}
	_ sdk.Msg = &MsgRemoveTokenPrice{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgRegisterTokenPriceQuery{}
	_ legacytx.LegacyMsg = &MsgRemoveTokenPrice{}
)

// ----------------------------------------------
//               MsgClaim
// ----------------------------------------------

func NewMsgRegisterTokenPriceQuery(sender, baseDenom, quoteDenom, poolId, osmosisBaseDenom, osmosisQuoteDenom string) *MsgRegisterTokenPriceQuery {
	return &MsgRegisterTokenPriceQuery{
		Sender:            sender,
		BaseDenom:         baseDenom,
		QuoteDenom:        quoteDenom,
		OsmosisPoolId:     poolId,
		OsmosisBaseDenom:  osmosisBaseDenom,
		OsmosisQuoteDenom: osmosisQuoteDenom,
	}
}

func (msg MsgRegisterTokenPriceQuery) Type() string {
	return TypeMsgRegisterTokenPriceQuery
}

func (msg MsgRegisterTokenPriceQuery) Route() string {
	return RouterKey
}

func (msg *MsgRegisterTokenPriceQuery) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgRegisterTokenPriceQuery) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterTokenPriceQuery) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.BaseDenom == "" {
		return errors.New("base-denom must be specified")
	}
	if msg.QuoteDenom == "" {
		return errors.New("quote-denom must be specified")
	}
	if _, err := strconv.ParseUint(msg.OsmosisPoolId, 10, 64); err != nil {
		return errors.New("osmosis-pool-id must be uint64")
	}
	if msg.OsmosisBaseDenom == "" {
		return errors.New("osmosis-base-denom must be specified")
	}
	if msg.OsmosisQuoteDenom == "" {
		return errors.New("osmosis-quote-denom must be specified")
	}

	return nil
}

// ----------------------------------------------
//               MsgRemoveTokenPrice
// ----------------------------------------------

func NewMsgRemoveTokenPrice(sender, baseDenom, quoteDenom, osmosisPoolId string) *MsgRemoveTokenPrice {
	return &MsgRemoveTokenPrice{
		Sender:        sender,
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: osmosisPoolId,
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
	if _, err := strconv.ParseUint(msg.OsmosisPoolId, 10, 64); err != nil {
		return errors.New("osmosis-pool-id must be uint64")
	}

	return nil
}
