package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v5/x/epochs/types"
)

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	epochs := []types.EpochInfo{
		{
			Identifier:              "day",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            sdkmath.ZeroInt(),
			CurrentEpochStartHeight: sdkmath.ZeroInt(),
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              "hour",
			StartTime:               time.Time{},
			Duration:                time.Hour,
			CurrentEpoch:            sdkmath.ZeroInt(),
			CurrentEpochStartHeight: sdkmath.ZeroInt(),
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}
	epochGenesis := types.NewGenesisState(epochs)

	bz, err := json.MarshalIndent(&epochGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	// TODO: Do some randomization later
	fmt.Printf("Selected deterministically generated epoch parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(epochGenesis)
}
