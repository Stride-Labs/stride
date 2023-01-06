package simulation

import (
	_ "fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/simulation"

	_ "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func ParamChanges(r *rand.Rand, cdc codec.Codec) []simtypes.ParamChange {
	// params := types.DefaultParams()
	// return []simtypes.ParamChange{
	// 	simulation.NewSimParamChange(types.ModuleName, fmt.Sprint(types.DefaultSafetyNumValidators),
	// 	func(r *rand.Rand) string {
	// 		return "TODO"
	// 	},
	// 	),
	// simulation.NewSimParamChange(types.ModuleName, string(types.KeyDelegateInterval),
	// 	func(r *rand.Rand) string {
	// 		return "TODO"
	// 	},
	// ),
	// }
	return []simtypes.ParamChange{} // TODO
}
