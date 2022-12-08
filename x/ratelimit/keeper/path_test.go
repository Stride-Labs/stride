package keeper_test

import (
	// "strconv"

	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
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

// Mock Channels:
//  stride (channel-0)   <-> cosmoshub (channel-100)
//  stride (channel-1)   <-> osmosis   (channel-200)
// osmosis (channel-300) <-> cosmoshub (channel-400)
var pathTestCases = []types.Path{
	// native ustrd on the stride -> cosmoshub channel
	{
		Id:         "ustrd/channel-0",
		TraceDenom: "ustrd",
		BaseDenom:  "ustrd",
		ChannelId:  "channel-0",
	},
	// uatom from cosmoshub -> stride: transfer/channel-0/uatom
	{
		Id:         "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		TraceDenom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		BaseDenom:  "uatom",
		ChannelId:  "channel-0",
	},
	// uosmo from osmosis -> stride: transfer/channel-1/uosmo
	{
		Id:         "ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B",
		TraceDenom: "ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B",
		BaseDenom:  "uatom",
		ChannelId:  "channel-1",
	},
	// uosmo from osmosis -> cosmoshub -> stride: transfer/channel-0/transfer/channel-400/uosmo
	{
		Id:         "ibc/",
		TraceDenom: "ibc/",
		BaseDenom:  "uosmo",
		ChannelId:  "channel-1",
	},
	// ustrd from stride -> osmosis -> stride: transfer/channel-200/transfer/channel-1
	{
		Id:         "ustrd/channel-1",
		TraceDenom: "ibc/",
		BaseDenom:  "ustrd",
		ChannelId:  "channel-1",
	},
}

func (s *KeeperTestSuite) TestFormatPath() {
	s.Require().Equal(keeper.FormatPathId("denom", "channel-0"), "denom_channel-0")
}

func (s *KeeperTestSuite) createPaths() {
	for _, path := range pathTestCases {
		s.App.RatelimitKeeper.SetPath(s.Ctx, path)
	}
}

func (s *KeeperTestSuite) TestAddPath() {
	// for _, path := range paths {
	// 	s.App.RatelimitKeeper.AddPath(s.Ctx, path.TraceDenom, path.ChannelId)
	// }
	// for _, path := range paths {
	// 	s.App.RatelimitKeeper.GetPath(s.Ctx, path.)
	// }
}

func (s *KeeperTestSuite) TestRemovePath() {
	s.createPaths()
	idToRemove := pathTestCases[0].Id

	s.App.RatelimitKeeper.RemovePath(s.Ctx, idToRemove)
	_, found := s.App.RatelimitKeeper.GetPath(s.Ctx, idToRemove)
	s.Require().False(found, "removed element found")
}

func (s *KeeperTestSuite) TestGetAllPaths() {
	s.createPaths()
	allPathsActual := s.App.RatelimitKeeper.GetAllPaths(s.Ctx)
	s.Require().ElementsMatch(pathTestCases, allPathsActual, "all paths")
}
