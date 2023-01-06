package simulation

import (
	_ "fmt"
	"math/rand"

	_ "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	// sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	// stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// Simulation operation weights constants
//
//nolint:gosec // these are not hardcoded credentials

const (
	// OpWeightMsgAddValidator                  = "op_weight_msg_add_validator"
	OpWeightMsgChangeValidatorWeight = "op_weight_msg_change_validator_weight"
	// OpWeightMsgClaimUndelegatedTokens        = "op_weight_msg_claim_undelegated_tokens"
	// OpWeightMsgDeleteValidator               = "op_weight_msg_delete_validator"
	// OpWeightMsgLiquidStake                   = "op_weight_msg_liquid_stake"
	// OpWeightMsgRebalanceValidators           = "op_weight_msg_rebalance_validators"
	// OpWeightMsgRestoreInterchainAccount      = "op_weight_msg_register_interchain_account"
	// OpWeightMsgUpdateValidatorSharesExchRate = "op_weight_msg_update_validator_shares_exch_rate"

	// DefaultWeightMsgAddValidator                  int = 100
	DefaultWeightMsgChangeValidatorWeight int = 100
	// DefaultWeightMsgClaimUndelegatedTokens        int = 100
	// DefaultWeightMsgDeleteValidator               int = 100
	// DefaultWeightMsgLiquidStake                   int = 100
	// DefaultWeightMsgRebalanceValidators           int = 100
	// DefaultWeightMsgRestoreInterchainAccount      int = 100
	// DefaultWeightMsgUpdateValidatorSharesExchRate int = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		// weightMsgAddValidator                  int
		weightMsgChangeValidatorWeight int
		// weightMsgClaimUndelegatedTokens        int
		// weightMsgDeleteValidator               int
		// weightMsgLiquidStake                   int
		// weightMsgRebalanceValidators           int
		// weightMsgRestoreInterchainAccount      int
		// weightMsgUpdateValidatorSharesExchRate int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgChangeValidatorWeight, &weightMsgChangeValidatorWeight, nil,
		func(_ *rand.Rand) {
			weightMsgChangeValidatorWeight = DefaultWeightMsgChangeValidatorWeight
		},
	)
	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgChangeValidatorWeight,
			SimulateMsgChangeValidatorWeight(ak, bk, k),
		)}
}
