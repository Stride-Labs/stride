// #nosec G101
package types

const (
	EventTypeLiquidStakeRequest        = "liquid_stake"
	EventTypeRedeemStakeRequest        = "redeem_stake"
	EventTypeConfirmUnbondedTokenSweep = "confirm_unbonded_token_sweep"

	AttributeKeyHostZone              = "host_zone"
	AttributeKeyLiquidStaker          = "liquid_staker"
	AttributeKeyNativeBaseDenom       = "native_base_denom"
	AttributeKeyNativeIBCDenom        = "native_ibc_denom"
	AttributeKeyNativeAmount          = "native_amount"
	AttributeKeyStTokenAmount         = "sttoken_amount"
	AttributeRecordId                 = "record_id"
	AttributeUndelegationNativeAmount = "undelegation_native_amount"
	AttributeTxHash                   = "tx_hash"
	AttributeSender                   = "sender"
)
