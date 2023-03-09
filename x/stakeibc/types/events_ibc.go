package types

// IBC events
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
	EventTypeHostZoneHalt       = "halt_zone"

	AttributeKeyConnectionId     = "connection_id"
	AttributeKeyRecipientChain   = "chain_id"
	AttributeKeyRecipientAddress = "recipient"
	AttributeKeyBurnAmount       = "burn_amount"
	AttributeKeyRedeemAmount     = "redeem_amount"
	AttributeKeySourceAddress    = "source"
	AttributeKeyHostZone         = "host_zone"
	AttributeKeyRedemptionRate   = "redemption_rate"

	AttributeValueCategory = ModuleName
)
