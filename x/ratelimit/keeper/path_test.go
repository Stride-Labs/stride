package keeper_test

// import (
// 	// "strconv"

// 	// ratelimittypes "github.com/Stride-Labs/stride/v4/x/ratelimit/types"
// )

import (
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
)

func (s *KeeperTestSuite) TestIBCDenom() {
	denom := "uosmo"
	for i := 0; i < 4; i++ {
		sourcePrefix := transfertypes.GetDenomPrefix("transfer", fmt.Sprintf("channel-%d", i))
		fmt.Println(sourcePrefix)
		prefixedDenom := sourcePrefix + denom

		fmt.Printf("%s\n", transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom())
	}
}

// func (s *KeeperTestSuite) createPaths() {
// 	pathIds := []string{
// 		"ustrd_channel-0",
// 		"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2_channel-0", // uatom channel-0
// 		""
// 	}
// 	paths := []ratelimittypes.Path{
// 		{
// 			Id: "ustrd-channel-0",
// 		},
// 		{
// 			Id: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
// 		},

// 	}
// 	paths := make([]ratelimittypes.Path, 5)
// 	for i := range paths {
// 		paths[i].Id = strconv.Itoa(i)
// 	}
// }

func (s *KeeperTestSuite) TestAddPath() {

}

func (s *KeeperTestSuite) TestSetPath() {

}

func (s *KeeperTestSuite) TestRemovePath() {

}

func (s *KeeperTestSuite) TestGetPath() {

}

func (s *KeeperTestSuite) TestGetAllPaths() {

}
