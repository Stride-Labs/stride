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
	EventTypeRegisterZone          = "register_zone"
	EventTypeRedemptionRequest     = "request_redemption"
	EventTypeLiquidStakeRequest    = "liquid_stake"
	EventTypeLSMLiquidStakeRequest = "lsm_liquid_stake"
	EventTypeLSMLiquidStakeFailed  = "lsm_liquid_stake_failed"
	EventTypeHostZoneHalt          = "halt_zone"

	AttributeKeyHostZone         = "host_zone"
	AttributeKeyConnectionId     = "connection_id"
	AttributeKeyRecipientChain   = "chain_id"
	AttributeKeyRecipientAddress = "recipient"
	AttributeKeyBurnAmount       = "burn_amount"
	AttributeKeyRedeemAmount     = "redeem_amount"
	AttributeKeySourceAddress    = "source"

	AttributeKeyRedemptionRate = "redemption_rate"

	AttributeKeyLiquidStaker      = "liquid_staker"
	AttributeKeyNativeBaseDenom   = "native_base_denom"
	AttributeKeyNativeIBCDenom    = "native_ibc_denom"
	AttributeKeyLSMTokenBaseDenom = "lsm_token_base_denom"
	AttributeKeyNativeAmount      = "native_amount"
	AttributeKeyStTokenAmount     = "sttoken_amount"
	AttributeKeyValidator         = "validator"

	AttributeKeyError = "error"

	AttributeValueCategory = ModuleName
)
