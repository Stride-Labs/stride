package types

// Events
const (
	EventTypeTimeout = "timeout"
	// this line is used by starport scaffolding # ibc/packet/event

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
)

const (
	EventTypeRegisterZone       = "register_zone"
	EventTypeRedemptionRequest  = "request_redemption"
	EventTypeLiquidStakeRequest = "liquid_stake"

	AttributeKeyConnectionId     = "connection_id"
	AttributeKeyRecipientChain   = "chain_id"
	AttributeKeyRecipientAddress = "recipient"
	AttributeKeyBurnAmount       = "burn_amount"
	AttributeKeyRedeemAmount     = "redeem_amount"
	AttributeKeySourceAddress    = "source"
	AttributeKeyUser             = "user"
	AttributeKeyHostZone         = "host_zone"
	AttributeKeyNativeBaseDenom  = "native_base_denom"
	AttributeKeyNativeIBCDenom   = "native_ibc_denom"
	AttributeKeyNativeAmount     = "native_amount"
	AttributeKeyStTokenAmount    = "sttoken_amount"

	AttributeValueCategory = ModuleName
)
