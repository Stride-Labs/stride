package keeper_test

import (
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestOnChanOpenAck() {

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
			owner := types.FormatICAAccountOwner(HostChainId, accountType)
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
		},
		TradeAccount: types.ICAAccount{
			ChainId: HostChainId,
		},
	}

	for _, accountType := range icaAccountTypes {
		// Reset the trade route for each test and wipe all addresses
		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, emptyTradeRoute)

		// Determine the port Id from the account type
		// If the portId is -1, pass a non-ica port
		portId := "not-ica-port"
		if accountType != -1 {
			owner := types.FormatICAAccountOwner(HostChainId, accountType)
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
	owner := types.FormatICAAccountOwner(HostChainId, accountType)
	portId, _ := icatypes.NewControllerPortID(owner)
	address := accountType.String()

	emptyTradeRoute.TradeAccount.ChainId = "different-chain-id"
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, emptyTradeRoute)

	err := s.App.StakeibcKeeper.StoreTradeRouteIcaAddress(s.Ctx, HostChainId, portId, address)
	s.Require().NoError(err, "no error expected when calling store trade route ICA for trade ICA with no chainId")

	s.checkTradeRouteAddressStored(-1) // checks no matches
}
