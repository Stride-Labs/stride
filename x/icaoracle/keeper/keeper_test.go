package keeper_test

import (
	"strconv"
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

var (
	HostChainId  = "host1"
	ConnectionId = "connection-0"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	QueryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.QueryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.ICAOracleKeeper)
}

// Helper function to create 5 oracle objects with various attributes
func (s *KeeperTestSuite) CreateTestOracles() []types.Oracle {
	oracles := []types.Oracle{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)

		channelId := "channel-" + suffix
		portId := "port-" + suffix

		oracle := types.Oracle{
			ChainId:         "chain-" + suffix,
			ConnectionId:    "connection-" + suffix,
			ChannelId:       channelId,
			PortId:          portId,
			IcaAddress:      "oracle-address",
			ContractAddress: "contract-address",
			Active:          true,
		}

		oracles = append(oracles, oracle)
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

		// Create open ICA channel
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portId, channelId, channeltypes.Channel{
			State: channeltypes.OPEN,
		})
	}
	return oracles
}
