package stakeibc

import (
	"math/rand"

	"github.com/Stride-Labs/stride/testutil/sample"
	stakeibcsimulation "github.com/Stride-Labs/stride/x/stakeibc/simulation"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = stakeibcsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgLiquidStake = "op_weight_msg_liquid_stake"
	// TODO: Determine the simulation weight value
	defaultWeightMsgLiquidStake int = 100

	opWeightMsgCreateEpochTracker = "op_weight_msg_epoch_tracker"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateEpochTracker int = 100

	opWeightMsgUpdateEpochTracker = "op_weight_msg_epoch_tracker"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateEpochTracker int = 100

	opWeightMsgDeleteEpochTracker = "op_weight_msg_epoch_tracker"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteEpochTracker int = 100

	opWeightMsgClaimUndelegatedTokens = "op_weight_msg_claim_undelegated_tokens"
	// TODO: Determine the simulation weight value
	defaultWeightMsgClaimUndelegatedTokens int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	stakeibcGenesis := types.GenesisState{
		Params:           types.DefaultParams(),
		PortId:           types.PortID,
		EpochTrackerList: []types.EpochTracker{},
	}
	// this line is used by starport scaffolding # simapp/module/genesisState
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&stakeibcGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgLiquidStake int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgLiquidStake, &weightMsgLiquidStake, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidStake = defaultWeightMsgLiquidStake
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgLiquidStake,
		stakeibcsimulation.SimulateMsgLiquidStake(am.accountKeeper, am.bankKeeper, am.keeper),
	))
	var weightMsgClaimUndelegatedTokens int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgClaimUndelegatedTokens, &weightMsgClaimUndelegatedTokens, nil,
		func(_ *rand.Rand) {
			weightMsgClaimUndelegatedTokens = defaultWeightMsgClaimUndelegatedTokens
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgClaimUndelegatedTokens,
		stakeibcsimulation.SimulateMsgClaimUndelegatedTokens(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
