package simulation


const (
	// Simulation operation weights constants
	OpWeightMsgAddValidator          			= "op_weight_msg_add_validator"
	OpWeightMsgChangeValidatorWeight   			= "op_weight_msg_change_validator_weight"
	OpWeightMsgClaimUndelegatedTokens 			= "op_weight_msg_claim_undelegated_tokens"
	OpWeightMsgDeleteValidator           		= "op_weight_msg_delete_validator"
	OpWeightMsgLiquidStake          			= "op_weight_msg_liquid_stake"
	OpWeightMsgRebalanceValidators           	= "op_weight_msg_rebalance_validators"
	OpWeightMsgRestoreInterchainAccount         = "op_weight_msg_restore_interchain_account"
	OpWeightMsgUpdateValidatorSharesExchRate    = "op_weight_msg_update_validator_shares_exch_rate"

	// Simulation default weights constants
	DefaultWeightMsgAddValidator          			= 100
	DefaultWeightMsgChangeValidatorWeight   		= 100
	DefaultWeightMsgClaimUndelegatedTokens 			= 100
	DefaultWeightMsgDeleteValidator           		= 100
	DefaultWeightMsgLiquidStake          			= 100
	DefaultWeightMsgRebalanceValidators           	= 100
	DefaultWeightMsgRestoreInterchainAccount      	= 100
	DefaultWeightMsgUpdateValidatorSharesExchRate 	= 100
)