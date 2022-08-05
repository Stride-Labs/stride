package stakeibc

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Stride-Labs/stride/testutil/sample"
	stakeibcsimulation "github.com/Stride-Labs/stride/x/stakeibc/simulation"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
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
	opWeightMsgRestoreInterchainAccount = "op_weight_msg_register_interchain_account" // #nosec
	// TODO: Determine the simulation weight value
	defaultWeightMsgRestoreInterchainAccount int = 100

	opWeightMsgQueryDelegation = "op_weight_msg_query_delegation"
	// TODO: Determine the simulation weight value
	defaultWeightMsgQueryDelegation int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	stakeibcGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
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

	var weightMsgRestoreInterchainAccount int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgRestoreInterchainAccount, &weightMsgRestoreInterchainAccount, nil,
		func(_ *rand.Rand) {
			weightMsgRestoreInterchainAccount = defaultWeightMsgRestoreInterchainAccount
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRestoreInterchainAccount,
		stakeibcsimulation.SimulateMsgRestoreInterchainAccount(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgQueryDelegation int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgQueryDelegation, &weightMsgQueryDelegation, nil,
		func(_ *rand.Rand) {
			weightMsgQueryDelegation = defaultWeightMsgQueryDelegation
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgQueryDelegation,
		stakeibcsimulation.SimulateMsgQueryDelegation(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
