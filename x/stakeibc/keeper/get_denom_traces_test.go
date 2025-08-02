package keeper_test

// // Note: this is for dockernet

// import (
// 	"fmt"

// 	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
// )

// func (s *KeeperTestSuite) TestIBCDenom() {
// 	chainId := "{CHAIN_ID}"
// 	denom := "{minimal_denom}"
// 	for i := 0; i < 4; i++ {
// 		sourcePrefix := transfertypes.GetDenomPrefix("transfer", fmt.Sprintf("channel-%d", i))
// 		prefixedDenom := sourcePrefix + denom

// 		fmt.Printf("IBC_%s_CHANNEL_%d_DENOM='%s'\n", chainId, i, transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom())
// 	}
// }
