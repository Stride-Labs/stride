package types

import (
	"errors"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/Stride-Labs/stride/v24/utils"
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

func NewMsgRegisterTokenPriceQuery(admin, baseDenom, quoteDenom, poolId, osmosisBaseDenom, osmosisQuoteDenom string) *MsgRegisterTokenPriceQuery {
	return &MsgRegisterTokenPriceQuery{
		Admin:             admin,
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
		msg.OsmosisPoolId,
		msg.OsmosisBaseDenom,
		msg.OsmosisQuoteDenom,
	)
}

// ----------------------------------------------
//               MsgRemoveTokenPriceQuery
// ----------------------------------------------

func NewMsgRemoveTokenPriceQuery(admin, baseDenom, quoteDenom, osmosisPoolId string) *MsgRemoveTokenPriceQuery {
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
	if _, err := strconv.ParseUint(msg.OsmosisPoolId, 10, 64); err != nil {
		return errors.New("osmosis-pool-id must be uint64")
	}

	return nil
}
