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
	EventTypeRegisterZone                = "register_zone"
	EventTypeRedemptionRequest           = "request_redemption"
	EventTypeLiquidStakeRequest          = "liquid_stake"
	EventTypeLSMLiquidStakeRequest       = "lsm_liquid_stake"
	EventTypeLSMLiquidStakeFailed        = "lsm_liquid_stake_failed"
	EventTypeHostZoneHalt                = "halt_zone"
	EventTypeValidatorExchangeRateChange = "validator_exchange_rate_change"
	EventTypeValidatorSlash              = "validator_slash"
	EventTypeUndelegation                = "undelegation"

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
	AttributeKeyTotalUnbondAmount = "total_unbond_amount"
	AttributeKeySweptAmount       = "swept_amount"
	AttributeKeyLSMTokenBaseDenom = "lsm_token_base_denom"
	AttributeKeyNativeAmount      = "native_amount"
	AttributeKeyStTokenAmount     = "sttoken_amount"
	AttributeKeyValidator         = "validator"

	AttributeKeyPreviousExchangeRate = "previous_exchange_rate"
	AttributeKeyCurrentExchangeRate  = "current_exchange_rate"
	AttributeKeySlashPercent         = "slash_percent"
	AttributeKeySlashAmount          = "slash_amount"
	AttributeKeyCurrentDelegation    = "current_delegation"

	AttributeKeyError = "error"

	AttributeValueCategory = ModuleName
)
