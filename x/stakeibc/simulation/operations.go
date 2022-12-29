package simulation

import (
	_"fmt"
	"math/rand"

	_"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	// sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	// stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)


// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simulation.WeightedOperations {
	var weightMsgAddValidator int
	appParams.GetOrGenerate(cdc, OpWeightMsgAddValidator, &weightMsgAddValidator, nil,
		func(_ *rand.Rand) {
			weightMsgAddValidator = DefaultWeightMsgAddValidator
		},
	)

	var weightMsgChangeValidatorWeight int
	appParams.GetOrGenerate(cdc, OpWeightMsgChangeValidatorWeight, &weightMsgChangeValidatorWeight, nil,
		func(_ *rand.Rand) {
			weightMsgChangeValidatorWeight = DefaultWeightMsgChangeValidatorWeight
		},
	)

	var weightMsgClaimUndelegatedTokens int
	appParams.GetOrGenerate(cdc, OpWeightMsgClaimUndelegatedTokens, &weightMsgClaimUndelegatedTokens, nil,
		func(_ *rand.Rand) {
			weightMsgClaimUndelegatedTokens = DefaultWeightMsgClaimUndelegatedTokens
		},
	)

	var weightMsgDeleteValidator int
	appParams.GetOrGenerate(cdc, OpWeightMsgDeleteValidator, &weightMsgDeleteValidator, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteValidator = DefaultWeightMsgDeleteValidator
		},
	)

	var weightMsgLiquidStake int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidStake, &weightMsgLiquidStake, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidStake = DefaultWeightMsgLiquidStake
		},
	)

	var weightMsgRebalanceValidators int
	appParams.GetOrGenerate(cdc, OpWeightMsgRebalanceValidators, &weightMsgRebalanceValidators, nil,
		func(_ *rand.Rand) {
			weightMsgRebalanceValidators = DefaultWeightMsgRebalanceValidators
		},
	)

	var weightMsgRestoreInterchainAccount int
	appParams.GetOrGenerate(cdc, OpWeightMsgRestoreInterchainAccount, &weightMsgRestoreInterchainAccount, nil,
		func(_ *rand.Rand) {
			weightMsgRestoreInterchainAccount = DefaultWeightMsgRestoreInterchainAccount
		},
	)

	var weightMsgUpdateValidatorSharesExchRate int
	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateValidatorSharesExchRate, &weightMsgUpdateValidatorSharesExchRate, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateValidatorSharesExchRate = DefaultWeightMsgUpdateValidatorSharesExchRate
		},
	)

	// stakeKeeper := sk.(stakingkeeper.Keeper)

	return simulation.WeightedOperations{
		// simulation.NewWeightedOperation(
		// 	weightMsgAddValidator,
		// 	SimulateMsgAddValidator(ak, bk, k),
		// ),
		simulation.NewWeightedOperation(
			weightMsgChangeValidatorWeight,
			SimulateMsgChangeValidatorWeight(ak, bk, k),
		),
		// simulation.NewWeightedOperation(
		// 	weightMsgWithdrawValidatorCommission,
		// 	SimulateMsgWithdrawValidatorCommission(ak, bk, k, stakeKeeper),
		// ),
		// simulation.NewWeightedOperation(
		// 	weightMsgFundCommunityPool,
		// 	SimulateMsgFundCommunityPool(ak, bk, k, stakeKeeper),
		// ),
	}
}