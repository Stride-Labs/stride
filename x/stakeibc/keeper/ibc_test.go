package keeper_test

import (
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// ------------------------------------------
//          	OnChanOpenAck
// ------------------------------------------

func (s *KeeperTestSuite) TestOnChanOpenAck() {
	// Define the mocked out ids for both the delegation and trade accounts
	delegationChainId := "delegation-1"
	delegationAddress := "delegation-address"
	delegationConnectionId := "connection-0"
	delegationChannelId := "channel-0"
	delegationClientId := "07-tendermint-0"

	tradeChainId := "trade-1"
	tradeAddress := "trade-address"
	tradeConnectionId := "connection-1"
	tradeChannelId := "channel-1"
	tradeClientId := "07-tendermint-1"

	// Create a host zone and trade route with existing ICA addresses
	// The ack should never overwrite them
	originalDelegationAddress := "original-delegation-address"
	originalTradeAddress := "original-trade-address"

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              delegationChainId,
		DelegationIcaAddress: originalDelegationAddress,
	})

	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		TradeAccount: types.ICAAccount{
			ChainId: tradeChainId,
			Type:    types.ICAAccountType_CONVERTER_TRADE,
			Address: originalTradeAddress,
		},
	})

	// Create the ICA channels for both the delegation and trade accounts
	delegationOwner := types.FormatHostZoneICAOwner(delegationChainId, types.ICAAccountType_DELEGATION)
	delegationPortId, _ := icatypes.NewControllerPortID(delegationOwner)

	tradeOwner := types.FormatTradeRouteICAOwner(tradeChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	tradePortId, _ := icatypes.NewControllerPortID(tradeOwner)

	// Mock out the relevant clients, channels, and ICA addresses so the callback can map back to the relevant info
	s.MockICAChannel(delegationConnectionId, delegationChannelId, delegationOwner, delegationAddress)
	s.MockICAChannel(tradeConnectionId, tradeChannelId, tradeOwner, tradeAddress)

	s.MockClientAndConnection(delegationChainId, delegationClientId, delegationConnectionId)
	s.MockClientAndConnection(tradeChainId, tradeClientId, tradeConnectionId)

	// Call the callback with the delegation ICA port and confirm the host zone is not modified
	err := s.App.StakeibcKeeper.OnChanOpenAck(s.Ctx, delegationPortId, delegationChannelId)
	s.Require().NoError(err, "no error expected when running callback with delegation port")

	hostZone := s.MustGetHostZone(delegationChainId)
	s.Require().Equal(originalDelegationAddress, hostZone.DelegationIcaAddress, "delegation address should not be overwritten")

	// Call the callback with the trade ICA port and confirm the trade route is not modified
	err = s.App.StakeibcKeeper.OnChanOpenAck(s.Ctx, tradePortId, tradeChannelId)
	s.Require().NoError(err, "no error expected when running callback with trade port")

	tradeRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been found")
	s.Require().Equal(originalTradeAddress, tradeRoute.TradeAccount.Address, "trade address should not be overwritten")
}

// ------------------------------------------
//         GetLightClientTime
// ------------------------------------------

type GetLightClientTestCase struct {
	connectionId              string
	expectedLightClientTime   int64
	expectedLightClientHeight int64
}

func (s *KeeperTestSuite) SetupGetLightClient() GetLightClientTestCase {
	connectionId := "connection-0"
	s.CreateTransferChannel("GAIA")

	// note this time is Jan 2020, set in the ibc test setup
	expectedLightClientTime := int64(1577923330000000000)
	// note this is the block height post-setup in the ibc test setup (creating connections, channels etc advances the block)
	// this may change as we amend the setup, please update accordingly!
	expectedLightClientHeight := int64(16)

	return GetLightClientTestCase{
		connectionId:              connectionId,
		expectedLightClientTime:   expectedLightClientTime,
		expectedLightClientHeight: expectedLightClientHeight,
	}
}

func (s *KeeperTestSuite) TestGetLightClientTime_Successful() {
	tc := s.SetupGetLightClient()

	actualLightClientTime, err := s.App.StakeibcKeeper.GetLightClientTime(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Greater(int(actualLightClientTime), 0, "light client time g.t. 0")
	s.Require().Equal(tc.expectedLightClientTime, int64(actualLightClientTime), "light client time matches expected time")

	// update LC to new block on host chain
	//   NOTE this advances the time!
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err, "update client")
	timeDelta := 10000000000

	actualLightClientTimeNewTime, err := s.App.StakeibcKeeper.GetLightClientTime(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "new light client time could be fetched")

	s.Require().Equal(int64(actualLightClientTimeNewTime), int64(actualLightClientTime+uint64(timeDelta)), "light client time increments by expected amount")
}

func (s *KeeperTestSuite) TestGetLightClient_InvalidConnection() {
	tc := s.SetupGetLightClient()
	tc.connectionId = "connection-invalid"

	_, err := s.App.StakeibcKeeper.GetLightClientTime(s.Ctx, tc.connectionId)
	s.Require().ErrorContains(err, "connection-id: connection-invalid: connection not found")

	_, err = s.App.StakeibcKeeper.GetLightClientHeight(s.Ctx, tc.connectionId)
	s.Require().ErrorContains(err, "connection-id: connection-invalid: connection not found")
}

func (s *KeeperTestSuite) TestGetLightClientHeight_Successful() {
	tc := s.SetupGetLightClient()

	actualLightClientHeight, err := s.App.StakeibcKeeper.GetLightClientHeight(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Greater(int(actualLightClientHeight), 0, "light client height g.t. 0")
	s.Require().Equal(tc.expectedLightClientHeight, int64(actualLightClientHeight), "light client height matches expected height")

	// update LC to new block on host chain
	//   NOTE this advances the block height!
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err, "update client")

	actualLightClientHeightNextBlock, err := s.App.StakeibcKeeper.GetLightClientHeight(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Equal(int64(actualLightClientHeight+1), int64(actualLightClientHeightNextBlock), "light client height matches expected height")
}
