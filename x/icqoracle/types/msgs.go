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
	TypeMsgUpdateParams            = "update_params"
)

var (
	_ sdk.Msg = &MsgRegisterTokenPriceQuery{}
	_ sdk.Msg = &MsgRemoveTokenPriceQuery{}
	_ sdk.Msg = &MsgUpdateParams{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgRegisterTokenPriceQuery{}
	_ legacytx.LegacyMsg = &MsgRemoveTokenPriceQuery{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

// ----------------------------------------------
//               MsgClaim
// ----------------------------------------------

func NewMsgRegisterTokenPriceQuery(
	admin string,
	baseDenom string,
	quoteDenom string,
	poolId uint64,
	osmosisBaseDenom string,
	osmosisQuoteDenom string,
) *MsgRegisterTokenPriceQuery {
	return &MsgRegisterTokenPriceQuery{
		Admin:             admin,
		BaseDenom:         baseDenom,
		QuoteDenom:        quoteDenom,
		OsmosisBaseDenom:  osmosisBaseDenom,
		OsmosisQuoteDenom: osmosisQuoteDenom,
		OsmosisPoolId:     poolId,
	}
}

func (msg MsgRegisterTokenPriceQuery) Type() string {
	return TypeMsgRegisterTokenPriceQuery
}

func (msg MsgRegisterTokenPriceQuery) Route() string {
	return RouterKey
}

func (msg *MsgRegisterTokenPriceQuery) GetSigners() []sdk.AccAddress {
	admin, _ := sdk.AccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{admin}
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
	admin, _ := sdk.AccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{admin}
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

// ----------------------------------------------
//               MsgUpdateParams
// ----------------------------------------------

func NewMsgUpdateParams(
	authority string,
	osmosisChainId string,
	osmosisConnectionId string,
	updateIntervalSec uint64,
	priceExpirationTimeoutSec uint64,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params: Params{
			OsmosisChainId:            osmosisChainId,
			OsmosisConnectionId:       osmosisConnectionId,
			UpdateIntervalSec:         updateIntervalSec,
			PriceExpirationTimeoutSec: priceExpirationTimeoutSec,
		},
	}
}

func (msg MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

func (msg MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if msg.Params.OsmosisChainId == "" {
		return errors.New("osmosis-chain-id must be specified")
	}
	if msg.Params.OsmosisConnectionId == "" {
		return errors.New("osmosis-connection-id must be specified")
	}
	if msg.Params.UpdateIntervalSec == 0 {
		return errors.New("update-interval-sec cannot be 0")
	}
	if msg.Params.PriceExpirationTimeoutSec == 0 {
		return errors.New("price-expiration-timeout-sec cannot be 0")
	}

	return nil
}
