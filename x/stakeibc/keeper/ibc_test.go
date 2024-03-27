package keeper_test

import (
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
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

	// Create a host zone with out any ICA addresses
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId: delegationChainId,
	})

	// Create a trade route without any ICA addresses
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		TradeAccount: types.ICAAccount{
			ChainId: tradeChainId,
			Type:    types.ICAAccountType_CONVERTER_TRADE,
		},
	})

	// Create the ICA channels for both the delegation and trade accounts
	delegationOwner := types.FormatHostZoneICAOwner(delegationChainId, types.ICAAccountType_DELEGATION)
	delegationPortId, _ := icatypes.NewControllerPortID(delegationOwner)

	tradeOwner := types.FormatTradeRouteICAOwner(tradeChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	tradePortId, _ := icatypes.NewControllerPortID(tradeOwner)

	// Mock out an ICA address for each
	s.App.ICAControllerKeeper.SetInterchainAccountAddress(s.Ctx, delegationConnectionId, delegationPortId, delegationAddress)
	s.App.ICAControllerKeeper.SetInterchainAccountAddress(s.Ctx, tradeConnectionId, tradePortId, tradeAddress)

	// Mock out a client and connection for each channel so the callback can map back from portId to chainId
	s.MockClientAndConnection(delegationChainId, delegationClientId, delegationConnectionId)
	s.MockClientAndConnection(tradeChainId, tradeClientId, tradeConnectionId)

	// Call the callback with the delegation ICA port and confirm the delegation address is set
	err := s.App.StakeibcKeeper.OnChanOpenAck(s.Ctx, delegationPortId, delegationChannelId)
	s.Require().NoError(err, "no error expected when running callback with delegation port")

	hostZone := s.MustGetHostZone(delegationChainId)
	s.Require().Equal(delegationAddress, hostZone.DelegationIcaAddress, "delegation address")

	// Call the callback with the trade ICA port and confirm the trade address is set
	err = s.App.StakeibcKeeper.OnChanOpenAck(s.Ctx, tradePortId, tradeChannelId)
	s.Require().NoError(err, "no error expected when running callback with trade port")

	tradeRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been round")
	s.Require().Equal(tradeAddress, tradeRoute.TradeAccount.Address, "trade address")

	// Call the callback with a non-ICA port and confirm the host zone and trade route remained unchanged
	err = s.App.StakeibcKeeper.OnChanOpenAck(s.Ctx, tradePortId, tradeChannelId)
	s.Require().NoError(err, "no error expected when running callback with non-ICA port")

	finalHostZone := s.MustGetHostZone(delegationChainId)
	s.Require().Equal(hostZone, finalHostZone, "host zone should not have been modified")

	finalTradeRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been round")
	s.Require().Equal(tradeRoute, finalTradeRoute, "trade route should not have been modified")
}

// ------------------------------------------
//      	StoreHostZoneIcaAddress
// ------------------------------------------

// Helper function to check that a single ICA address was stored on the host zone
// The address stored will match the string of the ICA account type
func (s *KeeperTestSuite) checkHostZoneAddressStored(accountType types.ICAAccountType) {
	// Determine the expected ICA addresses based on whether the account in question
	// is registered in this test case
	delegationAddress := ""
	if accountType == types.ICAAccountType_DELEGATION {
		delegationAddress = accountType.String()
	}
	withdrawalAddress := ""
	if accountType == types.ICAAccountType_WITHDRAWAL {
		withdrawalAddress = accountType.String()
	}
	redemptionAddress := ""
	if accountType == types.ICAAccountType_REDEMPTION {
		redemptionAddress = accountType.String()
	}
	feeAddress := ""
	if accountType == types.ICAAccountType_FEE {
		feeAddress = accountType.String()
	}
	communityPoolDepositAddress := ""
	if accountType == types.ICAAccountType_COMMUNITY_POOL_DEPOSIT {
		communityPoolDepositAddress = accountType.String()
	}
	communityPoolReturnAddress := ""
	if accountType == types.ICAAccountType_COMMUNITY_POOL_RETURN {
		communityPoolReturnAddress = accountType.String()
	}

	// Confirm the expected addresses with the host zone
	hostZone := s.MustGetHostZone(HostChainId)

	s.Require().Equal(delegationAddress, hostZone.DelegationIcaAddress, "delegation address")
	s.Require().Equal(withdrawalAddress, hostZone.WithdrawalIcaAddress, "withdrawal address")
	s.Require().Equal(redemptionAddress, hostZone.RedemptionIcaAddress, "redemption address")
	s.Require().Equal(feeAddress, hostZone.FeeIcaAddress, "fee address")
	s.Require().Equal(communityPoolDepositAddress, hostZone.CommunityPoolDepositIcaAddress, "community pool deposit address")
	s.Require().Equal(communityPoolReturnAddress, hostZone.CommunityPoolReturnIcaAddress, "commuity pool return address")
}

// Helper function to check that relevant ICA addresses are whitelisted after the callback
func (s *KeeperTestSuite) checkAddressesWhitelisted(accountType types.ICAAccountType) {
	if accountType == types.ICAAccountType_DELEGATION {
		isWhitelisted := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, DepositAddress, accountType.String())
		s.Require().True(isWhitelisted, "deposit -> delegation whitelist")
	}

	if accountType == types.ICAAccountType_FEE {
		sender := accountType.String()
		receiver := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.RewardCollectorName).GetAddress().String()

		isWhitelisted := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, sender, receiver)
		s.Require().True(isWhitelisted, "fee -> reward collector whitelist")
	}

	if accountType == types.ICAAccountType_COMMUNITY_POOL_DEPOSIT {
		sender := accountType.String()

		receiver := CommunityPoolStakeHoldingAddress
		isWhitelisted := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, sender, receiver)
		s.Require().True(isWhitelisted, "community pool deposit -> community pool stake holding")

		receiver = CommunityPoolRedeemHoldingAddress
		isWhitelisted = s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, sender, receiver)
		s.Require().True(isWhitelisted, "community pool deposit -> community pool redeem holding")
	}

	if accountType == types.ICAAccountType_COMMUNITY_POOL_RETURN {
		sender := CommunityPoolStakeHoldingAddress
		receiver := accountType.String()

		isWhitelisted := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, sender, receiver)
		s.Require().True(isWhitelisted, "community pool stake holding -> community pool return")
	}
}

func (s *KeeperTestSuite) TestStoreHostZoneIcaAddress() {
	// We'll run a test case for each ICA account, with two of them not being relevant for the host zone
	icaAccountTypes := []types.ICAAccountType{
		types.ICAAccountType_DELEGATION,
		types.ICAAccountType_WITHDRAWAL,
		types.ICAAccountType_REDEMPTION,
		types.ICAAccountType_FEE,
		types.ICAAccountType_COMMUNITY_POOL_DEPOSIT,
		types.ICAAccountType_COMMUNITY_POOL_RETURN,

		types.ICAAccountType_CONVERTER_TRADE, // not on the host zone
		-1,                                   // indicates test case for non-ICA port
	}

	for _, accountType := range icaAccountTypes {
		// Reset the host zone for each test and wipe all addresses
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
			ChainId:                           HostChainId,
			DepositAddress:                    DepositAddress,
			CommunityPoolStakeHoldingAddress:  CommunityPoolStakeHoldingAddress,
			CommunityPoolRedeemHoldingAddress: CommunityPoolRedeemHoldingAddress,
		})

		// Determine the port Id from the account type
		// If the portId is -1, pass a non-ica port
		portId := "not-ica-port"
		if accountType != -1 {
			owner := types.FormatHostZoneICAOwner(HostChainId, accountType)
			portId, _ = icatypes.NewControllerPortID(owner)
		}

		// Call StoreHostZoneIcaAddress with the portId
		// use the account name as the address to make the matching easier
		address := accountType.String()
		err := s.App.StakeibcKeeper.StoreHostZoneIcaAddress(s.Ctx, HostChainId, portId, address)
		s.Require().NoError(err, "no error expected when calling store host zone ICA for %s", accountType.String())

		// Check if the updated addresses matches expectations
		s.checkHostZoneAddressStored(accountType)

		// Check that the relevant accounts are white listed from the rate limiter
		s.checkAddressesWhitelisted(accountType)
	}
}

// ------------------------------------------
//      	StoreTradeRouteIcaAddress
// ------------------------------------------

// Helper function to check that a single ICA address was stored on the trade route
// The address stored will match the string of the ICA account type
func (s *KeeperTestSuite) checkTradeRouteAddressStored(accountType types.ICAAccountType) {
	// Determine the expected ICA addresses based on whether the account in question
	// is registered in this test case
	unwindAddress := ""
	if accountType == types.ICAAccountType_CONVERTER_UNWIND {
		unwindAddress = types.ICAAccountType_CONVERTER_UNWIND.String()
	}
	tradeAddress := ""
	if accountType == types.ICAAccountType_CONVERTER_TRADE {
		tradeAddress = types.ICAAccountType_CONVERTER_TRADE.String()
	}

	// Confirm the expected addresses with the host zone
	tradeRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been found")

	s.Require().Equal(unwindAddress, tradeRoute.RewardAccount.Address, "unwind address")
	s.Require().Equal(tradeAddress, tradeRoute.TradeAccount.Address, "trade address")
}

func (s *KeeperTestSuite) TestStoreTradeRouteIcaAddress() {
	// We'll run a test case for each the two ICA accounts, and 2 test cases for ports not on the trade route
	icaAccountTypes := []types.ICAAccountType{
		types.ICAAccountType_CONVERTER_UNWIND,
		types.ICAAccountType_CONVERTER_TRADE,

		types.ICAAccountType_DELEGATION, // not on the trade route
		-1,                              // indicates test case for non-ICA port
	}

	emptyTradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		RewardAccount: types.ICAAccount{
			ChainId: HostChainId,
			Type:    types.ICAAccountType_CONVERTER_UNWIND,
		},
		TradeAccount: types.ICAAccount{
			ChainId: HostChainId,
			Type:    types.ICAAccountType_CONVERTER_TRADE,
		},
	}

	for _, accountType := range icaAccountTypes {
		// Reset the trade route for each test and wipe all addresses
		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, emptyTradeRoute)

		// Determine the port Id from the account type
		// If the portId is -1, pass a non-ica port
		portId := "not-ica-port"
		if accountType != -1 {
			owner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, accountType)
			portId, _ = icatypes.NewControllerPortID(owner)
		}

		// Call StoreTradeRouteIcaAddress with the portId
		// use the account name as the address to make the matching easier
		address := accountType.String()
		err := s.App.StakeibcKeeper.StoreTradeRouteIcaAddress(s.Ctx, HostChainId, portId, address)
		s.Require().NoError(err, "no error expected when calling store trade route ICA for %s", accountType.String())

		// Check if the updated addresses matches expectations
		s.checkTradeRouteAddressStored(accountType)
	}

	// Check with a matching port, but no matching chainId
	accountType := types.ICAAccountType_CONVERTER_TRADE
	owner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, accountType)
	portId, _ := icatypes.NewControllerPortID(owner)
	address := accountType.String()

	emptyTradeRoute.TradeAccount.ChainId = "different-chain-id"
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, emptyTradeRoute)

	err := s.App.StakeibcKeeper.StoreTradeRouteIcaAddress(s.Ctx, HostChainId, portId, address)
	s.Require().NoError(err, "no error expected when calling store trade route ICA for trade ICA with no chainId")

	s.checkTradeRouteAddressStored(-1) // checks no matches
}

// ------------------------------------------
//         GetLightClientTimeSafely
// ------------------------------------------

type GetLightClientSafelyTestCase struct {
	connectionId              string
	expectedLightClientTime   int64
	expectedLightClientHeight int64
}

func (s *KeeperTestSuite) SetupGetLightClientSafely() GetLightClientSafelyTestCase {
	connectionId := "connection-0"
	s.CreateTransferChannel("GAIA")

	// note this time is Jan 2020, set in the ibc test setup
	expectedLightClientTime := int64(1577923355000000000)
	// note this is the block height post-setup in the ibc test setup (creating connections, channels etc advances the block)
	//        this may change as we ament the setup, please update accordingly!
	expectedLightClientHeight := int64(17)

	return GetLightClientSafelyTestCase{
		connectionId:              connectionId,
		expectedLightClientTime:   expectedLightClientTime,
		expectedLightClientHeight: expectedLightClientHeight,
	}
}

func (s *KeeperTestSuite) TestGetLightClientTimeSafely_Successful() {
	tc := s.SetupGetLightClientSafely()

	actualLightClientTime, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Greater(int(actualLightClientTime), 0, "light client time g.t. 0")
	s.Require().Equal(tc.expectedLightClientTime, int64(actualLightClientTime), "light client time matches expected time")

	// update LC to new block on host chain
	//   NOTE this advances the time!
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err, "update client")
	timeDelta := 10000000000

	actualLightClientTimeNewTime, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "new light client time could be fetched")

	s.Require().Equal(int64(actualLightClientTimeNewTime), int64(actualLightClientTime+uint64(timeDelta)), "light client time increments by expected amount")
}

func (s *KeeperTestSuite) TestGetLightClientSafely_InvalidConnection() {
	tc := s.SetupGetLightClientSafely()
	tc.connectionId = "connection-invalid"

	_, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().ErrorContains(err, "invalid connection id", "get lc time: error complains about invalid connection id")

	_, err = s.App.StakeibcKeeper.GetLightClientHeightSafely(s.Ctx, tc.connectionId)
	s.Require().ErrorContains(err, "invalid connection id", "get lc height: error complains about invalid connection id")
}

func (s *KeeperTestSuite) TestGetLightClientHeightSafely_Successful() {
	tc := s.SetupGetLightClientSafely()

	actualLightClientHeight, err := s.App.StakeibcKeeper.GetLightClientHeightSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Greater(int(actualLightClientHeight), 0, "light client height g.t. 0")
	s.Require().Equal(int64(actualLightClientHeight), tc.expectedLightClientHeight, "light client height matches expected height")

	// update LC to new block on host chain
	//   NOTE this advances the block height!
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err, "update client")

	actualLightClientHeightNextBlock, err := s.App.StakeibcKeeper.GetLightClientHeightSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Equal(int64(actualLightClientHeightNextBlock), int64(actualLightClientHeight+1), "light client height matches expected height")
}
