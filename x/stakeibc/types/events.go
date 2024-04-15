// #nosec G101
package types

// Events
const (
	EventTypeTimeout = "timeout"

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
)

const (
	EventTypeRegisterZone                      = "register_zone"
	EventTypeRedemptionRequest                 = "request_redemption"
	EventTypeLiquidStakeRequest                = "liquid_stake"
	EventTypeRedeemStakeRequest                = "redeem_stake"
	EventTypeLSMLiquidStakeRequest             = "lsm_liquid_stake"
	EventTypeHostZoneHalt                      = "halt_zone"
	EventTypeValidatorSharesToTokensRateChange = "validator_shares_to_tokens_rate_change"
	EventTypeValidatorSlash                    = "validator_slash"
	EventTypeUndelegation                      = "undelegation"

	AttributeKeyHostZone         = "host_zone"
	AttributeKeyConnectionId     = "connection_id"
	AttributeKeyRecipientChain   = "chain_id"
	AttributeKeyRecipientAddress = "recipient"
	AttributeKeyBurnAmount       = "burn_amount"
	AttributeKeyRedeemAmount     = "redeem_amount"
	AttributeKeySourceAddress    = "source"

	AttributeKeyRedemptionRate = "redemption_rate"

	AttributeKeyLiquidStaker       = "liquid_staker"
	AttributeKeyRedeemer           = "redeemer"
	AttributeKeyReceiver           = "receiver"
	AttributeKeyNativeBaseDenom    = "native_base_denom"
	AttributeKeyNativeIBCDenom     = "native_ibc_denom"
	AttributeKeyTotalUnbondAmount  = "total_unbond_amount"
	AttributeKeyLSMTokenBaseDenom  = "lsm_token_base_denom" // #nosec G101
	AttributeKeyNativeAmount       = "native_amount"
	AttributeKeyStTokenAmount      = "sttoken_amount"
	AttributeKeyValidator          = "validator"
	AttributeKeyTransactionStatus  = "transaction_status"
	AttributeKeyLSMLiquidStakeTxId = "lsm_liquid_stake_tx_id"

	AttributeKeyPreviousSharesToTokensRate = "previous_shares_to_tokens_rate"
	AttributeKeyCurrentSharesToTokensRate  = "current_shares_to_tokens_rate"
	AttributeKeySlashPercent               = "slash_percent"
	AttributeKeySlashAmount                = "slash_amount"
	AttributeKeyCurrentDelegation          = "current_delegation"

	AttributeKeyError = "error"

	AttributeValueCategory             = ModuleName
	AttributeValueTransactionSucceeded = "success"
	AttributeValueTransactionPending   = "pending"
	AttributeValueTransactionFailed    = "failed"
)
