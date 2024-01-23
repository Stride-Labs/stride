// #nosec G101
package types

const (
	EventTypeLiquidStakeRequest        = "liquid_stake"
	EventTypeRedeemStakeRequest        = "redeem_stake"
	EventTypeConfirmDelegationResponse = "confirm_delegation"
	EventTypeConfirmUnbondedTokenSweep = "confirm_unbonded_token_sweep"
	EventTypeConfirmUndelegation       = "confirm_undelegation"

	AttributeKeyHostZone              = "host_zone"
	AttributeKeyLiquidStaker          = "liquid_staker"
	AttributeKeyNativeBaseDenom       = "native_base_denom"
	AttributeKeyNativeIBCDenom        = "native_ibc_denom"
	AttributeKeyNativeAmount          = "native_amount"
	AttributeKeyStTokenAmount         = "sttoken_amount"
	AttributeRecordId                 = "record_id"
	AttributeDelegationNativeAmount   = "delegation_native_amount"
	AttributeUndelegationNativeAmount = "undelegation_native_amount"
	AttributeTxHash                   = "tx_hash"
	AttributeSender                   = "sender"
)
