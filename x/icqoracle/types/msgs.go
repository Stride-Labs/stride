package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/Stride-Labs/stride/v25/utils"
)

const (
	TypeMsgRegisterTokenPriceQuery = "register_token_price_query"
	TypeMsgRemoveTokenPriceQuery   = "remove_token_price_query"
)

var (
	_ sdk.Msg = &MsgRegisterTokenPriceQuery{}
	_ sdk.Msg = &MsgRemoveTokenPriceQuery{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgRegisterTokenPriceQuery{}
	_ legacytx.LegacyMsg = &MsgRemoveTokenPriceQuery{}
)

// ----------------------------------------------
//               MsgClaim
// ----------------------------------------------

func NewMsgRegisterTokenPriceQuery(
	admin string,
	baseDenom string,
	quoteDenom string,
	baseDecimals int64,
	quoteDecimals int64,
	poolId uint64,
	osmosisBaseDenom string,
	osmosisQuoteDenom string,
) *MsgRegisterTokenPriceQuery {
	return &MsgRegisterTokenPriceQuery{
		Admin:              admin,
		BaseDenom:          baseDenom,
		QuoteDenom:         quoteDenom,
		BaseDenomDecimals:  baseDecimals,
		QuoteDenomDecimals: quoteDecimals,
		OsmosisBaseDenom:   osmosisBaseDenom,
		OsmosisQuoteDenom:  osmosisQuoteDenom,
		OsmosisPoolId:      poolId,
	}
}

func (msg MsgRegisterTokenPriceQuery) Type() string {
	return TypeMsgRegisterTokenPriceQuery
}

func (msg MsgRegisterTokenPriceQuery) Route() string {
	return RouterKey
}

func (msg *MsgRegisterTokenPriceQuery) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Admin)
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
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}
	return ValidateTokenPriceQueryParams(
		msg.BaseDenom,
		msg.QuoteDenom,
		msg.BaseDenomDecimals,
		msg.QuoteDenomDecimals,
		msg.OsmosisPoolId,
		msg.OsmosisBaseDenom,
		msg.OsmosisQuoteDenom,
	)
}

// ----------------------------------------------
//               MsgRemoveTokenPriceQuery
// ----------------------------------------------

func NewMsgRemoveTokenPriceQuery(admin, baseDenom, quoteDenom string, osmosisPoolId uint64) *MsgRemoveTokenPriceQuery {
	return &MsgRemoveTokenPriceQuery{
		Admin:         admin,
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: osmosisPoolId,
	}
}

func (msg MsgRemoveTokenPriceQuery) Type() string {
	return TypeMsgRemoveTokenPriceQuery
}

func (msg MsgRemoveTokenPriceQuery) Route() string {
	return RouterKey
}

func (msg *MsgRemoveTokenPriceQuery) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgRemoveTokenPriceQuery) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveTokenPriceQuery) ValidateBasic() error {
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}
	if msg.BaseDenom == "" {
		return errors.New("base-denom must be specified")
	}
	if msg.QuoteDenom == "" {
		return errors.New("quote-denom must be specified")
	}
	if msg.OsmosisPoolId == 0 {
		return errors.New("osmosis-pool-id must be specified")
	}

	return nil
}
