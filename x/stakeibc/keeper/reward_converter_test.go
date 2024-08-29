package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v24/utils"
	epochtypes "github.com/Stride-Labs/stride/v24/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v24/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v24/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v24/x/stakeibc/types"
)

// Useful across all balance query icqcallback tests
type BalanceQueryCallbackTestCase struct {
	TradeRoute types.TradeRoute
	Response   ICQCallbackArgs
	Balance    sdkmath.Int
	ChannelID  string
	PortID     string
}

type TransferRewardHostToTradeTestCase struct {
	TradeRoute          types.TradeRoute
	TransferAmount      sdkmath.Int
	ExpectedTransferMsg transfertypes.MsgTransfer
	ChannelID           string
	PortID              string
}

// --------------------------------------------------------------
//                    CalculateRewardsSplit
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestCalculateRewardsSplit() {
	testCases := []struct {
		name                     string
		communityPoolLiquidStake sdkmath.Int
		totalStTokenSupply       sdkmath.Int
		rewardAmount             sdkmath.Int
		strideFee                uint64
		rebateRate               sdk.Dec
		expectedRebateAmount     sdkmath.Int
		expectedStrideFeeAmount  sdkmath.Int
		expectedReinvestAmount   sdkmath.Int
		expectedError            string
	}{
		{
			// 10 CP Liquid Stake, 100 TVL => 10% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 10% contribution * 50% rebate => 5 rebate
			// 100 total fees - 5 rebate => 95 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "case 1",
			communityPoolLiquidStake: sdkmath.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.5"),

			expectedRebateAmount:    sdkmath.NewInt(5),
			expectedStrideFeeAmount: sdkmath.NewInt(95),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// (Example #1 but with a 2x bigger liquid stake)
			// 20 CP Liquid Stake, 100 TVL => 20% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 20% contribution * 50% rebate => 10 rebate
			// 100 total fees - 10 rebate => 90 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "case 2",
			communityPoolLiquidStake: sdkmath.NewInt(20),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.5"),

			expectedRebateAmount:    sdkmath.NewInt(10),
			expectedStrideFeeAmount: sdkmath.NewInt(90),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// (Example #1 but with a 2x larger TVL)
			// 10 CP Liquid Stake, 200 TVL => 5% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 5% contribution * 50% rebate => 2.5 rebate (truncated to 2)
			// 100 total fees - 2 rebate => 98 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "case 3",
			communityPoolLiquidStake: sdkmath.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(200),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.5"),

			expectedRebateAmount:    sdkmath.NewInt(2),
			expectedStrideFeeAmount: sdkmath.NewInt(98),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// (Example #1 but with a 2x larger stride fee)
			// 10 CP Liquid Stake, 100 TVL => 10% contribution
			// 1000 rewards, 20% stride fee => 200 total fees
			// 200 total fees * 10% contribution * 50% rebate => 10 rebate
			// 200 total fees - 10 rebate => 190 stride fee
			// 1000 rewards - 200 total fees => 800 reinvested
			name:                     "case 4",
			communityPoolLiquidStake: sdkmath.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                20,
			rebateRate:               sdk.MustNewDecFromStr("0.5"),

			expectedRebateAmount:    sdkmath.NewInt(10),
			expectedStrideFeeAmount: sdkmath.NewInt(190),
			expectedReinvestAmount:  sdkmath.NewInt(800),
		},
		{
			// (Example #1 but with a smaller rebate)
			// 10 CP Liquid Stake, 100 TVL => 10% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 10% contribution * 20% rebate => 2 rebate
			// 100 total fees - 2 rebate => 98 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "case 5",
			communityPoolLiquidStake: sdkmath.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.2"),

			expectedRebateAmount:    sdkmath.NewInt(2),
			expectedStrideFeeAmount: sdkmath.NewInt(98),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// (Example #1 but with a larger rebate)
			// 10 CP Liquid Stake, 100 TVL => 10% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 10% contribution * 79% rebate => 7.9 rebate (truncated to 7)
			// 100 total fees - 2 rebate => 98 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "case 6",
			communityPoolLiquidStake: sdkmath.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.79"),

			expectedRebateAmount:    sdkmath.NewInt(7),
			expectedStrideFeeAmount: sdkmath.NewInt(93),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// No rebate
			// 10% fees off 1000 rewards = 100 stride fees, 900 reinvest
			name:               "nil rebate",
			totalStTokenSupply: sdkmath.NewInt(100),
			rewardAmount:       sdkmath.NewInt(1000),
			strideFee:          10,

			expectedRebateAmount:    sdkmath.NewInt(0),
			expectedStrideFeeAmount: sdkmath.NewInt(100),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// 0% rebate - all fees go to stride
			// 10% fees off 1000 rewards = 100 stride fees, 900 reinvest
			name:               "zero rebate",
			totalStTokenSupply: sdkmath.NewInt(100),
			rewardAmount:       sdkmath.NewInt(1000),
			strideFee:          10,
			rebateRate:         sdk.ZeroDec(),

			expectedRebateAmount:    sdkmath.NewInt(0),
			expectedStrideFeeAmount: sdkmath.NewInt(100),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// 100% rebate
			// 10 CP Liquid Stake, 100 TVL => 10% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 10% contribution * 100% rebate => 10 rebate
			// 100 total fees - 10 rebate => 90 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "full rebate",
			communityPoolLiquidStake: sdkmath.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.OneDec(),

			expectedRebateAmount:    sdkmath.NewInt(10),
			expectedStrideFeeAmount: sdkmath.NewInt(90),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// Liquid staked amount 0 - effectively the same as no rebate
			// 10% fees off 1000 rewards = 100 stride fees, 900 reinvest
			name:                     "zero liquid staked",
			communityPoolLiquidStake: sdkmath.NewInt(0),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.50"), // ignored since 0 LS'd

			expectedRebateAmount:    sdkmath.NewInt(0),
			expectedStrideFeeAmount: sdkmath.NewInt(100),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// Liquid stake represents all of TVL
			// Community pool liquid stake represents full TVL
			// 100 CP Liquid Stake, 100 TVL => 100% contribution
			// 1000 rewards, 10% stride fee => 100 total fees
			// 100 total fees * 100% contribution * 50% rebate => 50 rebate
			// 100 total fees - 50 rebate => 50 stride fee
			// 1000 rewards - 100 total fees => 900 reinvested
			name:                     "liquid stake represents full TVL",
			communityPoolLiquidStake: sdkmath.NewInt(100),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.50"),

			expectedRebateAmount:    sdkmath.NewInt(50),
			expectedStrideFeeAmount: sdkmath.NewInt(50),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// 100% contribution, 100% rebate
			// Community pool gets all fees
			// 10% fees off 1000 rewards = 100 rebate, 900 reinvest
			name:                     "liquid stake represents full TVL and full rebate",
			communityPoolLiquidStake: sdkmath.NewInt(100),
			totalStTokenSupply:       sdkmath.NewInt(100),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.OneDec(),

			expectedRebateAmount:    sdkmath.NewInt(100),
			expectedStrideFeeAmount: sdkmath.NewInt(0),
			expectedReinvestAmount:  sdkmath.NewInt(900),
		},
		{
			// No tvl - should error
			name:                     "no tvl",
			communityPoolLiquidStake: sdk.NewInt(10),
			totalStTokenSupply:       sdkmath.NewInt(0),
			rewardAmount:             sdkmath.NewInt(1000),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.5"),

			expectedError: "unable to calculate rebate amount",
		},
		{
			// Liquid staked amount is greater than the TVL - should error
			name:                     "liquid staked more than tvl",
			communityPoolLiquidStake: sdk.NewInt(1001),
			totalStTokenSupply:       sdkmath.NewInt(1000),
			rewardAmount:             sdkmath.NewInt(100),
			strideFee:                10,
			rebateRate:               sdk.MustNewDecFromStr("0.5"),

			expectedError: "community pool liquid staked amount greater than total delegations",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // resets supply

			// Build out the host zone - only add the rebate struct if one of the rebate fields was provided
			hostZone := types.HostZone{
				ChainId:   chainId,
				HostDenom: HostDenom,
			}
			if !tc.communityPoolLiquidStake.IsNil() {
				hostZone.CommunityPoolRebate = &types.CommunityPoolRebate{
					RebateRate:                tc.rebateRate,
					LiquidStakedStTokenAmount: tc.communityPoolLiquidStake,
				}
			}

			// Store the fee as a param
			params := types.DefaultParams()
			params.StrideCommission = tc.strideFee
			s.App.StakeibcKeeper.SetParams(s.Ctx, params)

			// Mint stTokens to populate the supply
			stCoin := sdk.NewCoin(utils.StAssetDenomFromHostZoneDenom(HostDenom), tc.totalStTokenSupply)
			s.FundAccount(s.TestAccs[0], stCoin)

			// Call the tested function to get the expected amounts
			rewardsSplit, actualError := s.App.StakeibcKeeper.CalculateRewardsSplit(
				s.Ctx,
				hostZone,
				tc.rewardAmount,
			)

			// Confirm the amounts and error
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError, "error expected")
			} else {
				s.Require().Equal(tc.expectedRebateAmount.Int64(), rewardsSplit.RebateAmount.Int64(), "rebate amount")
				s.Require().Equal(tc.expectedStrideFeeAmount.Int64(), rewardsSplit.StrideFeeAmount.Int64(), "stride fee amount")
				s.Require().Equal(tc.expectedReinvestAmount.Int64(), rewardsSplit.ReinvestAmount.Int64(), "reinvest amount")
			}
		})
	}
}

// --------------------------------------------------------------
//                   BuildTradeAuthzMsg
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestBuildTradeAuthzMsg() {
	granterAddress := "trade_ica"
	granteeAddress := "trade_controller"

	tradeRoute := types.TradeRoute{
		TradeAccount: types.ICAAccount{
			Address: granterAddress,
		},
	}

	testCases := map[bool]string{
		false: "/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn",
		true:  "/osmosis.gamm.v1beta1.MsgSwapExactAmountIn",
	}

	for legacy, expectedTypeUrl := range testCases {
		// Test granting trade permissions
		msgs, err := s.App.StakeibcKeeper.BuildTradeAuthzMsg(
			s.Ctx,
			tradeRoute,
			types.AuthzPermissionChange_GRANT,
			granteeAddress,
			legacy,
		)
		s.Require().NoError(err, "no error expected when building grant message")
		s.Require().Len(msgs, 1, "there should be one message")

		grantMsg, ok := msgs[0].(*authz.MsgGrant)
		s.Require().True(ok, "message should be of type grant")
		s.Require().Equal(granterAddress, grantMsg.Granter, "granter of grant message")
		s.Require().Equal(granteeAddress, grantMsg.Grantee, "grantee of grant message")

		authorization, err := grantMsg.Grant.GetAuthorization()
		expectedExpiration := s.Ctx.BlockTime().Add(time.Hour * 24 * 365 * 100)
		s.Require().NoError(err)
		s.Require().Equal(expectedTypeUrl, authorization.MsgTypeURL(), "grant msg type url")
		s.Require().Equal(expectedExpiration, *grantMsg.Grant.Expiration, "expiration should be one year from the current block time")

		// Test revoking trade permissions
		msgs, err = s.App.StakeibcKeeper.BuildTradeAuthzMsg(
			s.Ctx,
			tradeRoute,
			types.AuthzPermissionChange_REVOKE,
			granteeAddress,
			legacy,
		)
		s.Require().NoError(err, "no error expected when building revoke message")
		s.Require().Len(msgs, 1, "there should be one message")

		revokeMsg, ok := msgs[0].(*authz.MsgRevoke)
		s.Require().True(ok, "message should be of type revoke")
		s.Require().Equal(granterAddress, revokeMsg.Granter, "granter of revoke message")
		s.Require().Equal(granteeAddress, revokeMsg.Grantee, "grantee of revoke message")
		s.Require().Equal(expectedTypeUrl, revokeMsg.MsgTypeUrl, "revoke msg type url")

		// Test invalid permissions
		_, err = s.App.StakeibcKeeper.BuildTradeAuthzMsg(s.Ctx, tradeRoute, 100, granteeAddress, legacy)
		s.Require().ErrorContains(err, "invalid permission change")
	}

}

// --------------------------------------------------------------
//                   Transfer Host to Trade
// --------------------------------------------------------------

func (s *KeeperTestSuite) SetupTransferRewardTokensHostToTradeTestCase() TransferRewardHostToTradeTestCase {
	// Create an ICA channel for the transfer submission
	owner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)
	channelId, portId := s.CreateICAChannel(owner)

	// Define components of transfer message
	hostToRewardChannelId := "channel-0"
	rewardToTradeChannelId := "channel-1"

	rewardDenomOnHostZone := "ibc/reward_on_host"
	rewardDenomOnRewardZone := RewardDenom

	withdrawalAddress := "withdrawal_address"
	unwindAddress := "unwind_address"
	tradeAddress := "trade_address"

	transferAmount := sdk.NewInt(1000)
	transferToken := sdk.NewCoin(rewardDenomOnHostZone, transferAmount)
	minTransferAmount := sdk.NewInt(500)

	currentTime := s.Ctx.BlockTime()
	epochLength := time.Second * 10                               // 10 seconds
	transfer1TimeoutTimestamp := currentTime.Add(time.Second * 5) // 5 seconds from now (halfway through)
	transfer2TimeoutDuration := "5s"

	// Create a trade route with the relevant addresses and transfer channels
	route := types.TradeRoute{
		HostToRewardChannelId:  hostToRewardChannelId,
		RewardToTradeChannelId: rewardToTradeChannelId,

		RewardDenomOnHostZone:   rewardDenomOnHostZone,
		RewardDenomOnRewardZone: rewardDenomOnRewardZone,
		HostDenomOnHostZone:     HostDenom,

		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Address:      withdrawalAddress,
			ConnectionId: ibctesting.FirstConnectionID,
			Type:         types.ICAAccountType_WITHDRAWAL,
		},
		RewardAccount: types.ICAAccount{
			Address: unwindAddress,
		},
		TradeAccount: types.ICAAccount{
			Address: tradeAddress,
		},

		MinTransferAmount: minTransferAmount,
	}

	// Create an epoch tracker to dictate the timeout
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, epochLength)

	// Define the expected transfer message using all the above
	memoJSON := fmt.Sprintf(`{"forward":{"receiver":"%s","port":"transfer","channel":"%s","timeout":"%s","retries":0}}`,
		tradeAddress, rewardToTradeChannelId, transfer2TimeoutDuration)

	expectedMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostToRewardChannelId,
		Token:            transferToken,
		Sender:           withdrawalAddress,
		Receiver:         unwindAddress,
		TimeoutTimestamp: uint64(transfer1TimeoutTimestamp.UnixNano()),
		Memo:             memoJSON,
	}

	return TransferRewardHostToTradeTestCase{
		TradeRoute:          route,
		TransferAmount:      transferAmount,
		ExpectedTransferMsg: expectedMsg,
		ChannelID:           channelId,
		PortID:              portId,
	}
}

func (s *KeeperTestSuite) TestBuildHostToTradeTransferMsg_Success() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Confirm the generated message matches expectations
	actualMsg, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	s.Require().NoError(err, "no error expected when building transfer message")
	s.Require().Equal(tc.ExpectedTransferMsg, actualMsg, "transfer message should have matched")
}

func (s *KeeperTestSuite) TestBuildHostToTradeTransferMsg_InvalidICAAddress() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Check unregisted ICA addresses cause failures
	invalidRoute := tc.TradeRoute
	invalidRoute.HostAccount.Address = ""
	_, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no host account found")

	invalidRoute = tc.TradeRoute
	invalidRoute.RewardAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no reward account found")

	invalidRoute = tc.TradeRoute
	invalidRoute.TradeAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no trade account found")
}

func (s *KeeperTestSuite) TestBuildHostToTradeTransferMsg_EpochNotFound() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Delete the epoch tracker and confirm the message cannot be built
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	s.Require().ErrorContains(err, "epoch not found")
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_Success() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Check that the transfer ICA is submitted when the function is called
	s.CheckICATxSubmitted(tc.PortID, tc.ChannelID, func() error {
		return s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	})
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_TransferAmountBelowMin() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Attempt to call the function with an transfer amount below the min,
	// it should not submit an ICA
	invalidTransferAmount := tc.TradeRoute.MinTransferAmount.Sub(sdkmath.OneInt())
	s.CheckICATxNotSubmitted(tc.PortID, tc.ChannelID, func() error {
		return s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, invalidTransferAmount, tc.TradeRoute)
	})
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_FailedToSubmitICA() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Remove the connection ID and confirm the ICA submission fails
	invalidRoute := tc.TradeRoute
	invalidRoute.HostAccount.ConnectionId = ""

	err := s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "Failed to submit ICA tx")
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_EpochNotFound() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Delete the epoch tracker and confirm the transfer cannot be initiated
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	s.Require().ErrorContains(err, "epoch not found")
}

// --------------------------------------------------------------
//                   Transfer Trade to Trade
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestTransferConvertedTokensTradeToHost() {
	transferAmount := sdkmath.NewInt(1000)

	// Register a trade ICA account for the transfer
	owner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	channelId, portId := s.CreateICAChannel(owner)

	// Create trade route with fields needed for transfer
	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,

		HostDenomOnTradeZone: "ibc/host-on-trade",
		TradeToHostChannelId: "channel-1",
		HostAccount: types.ICAAccount{
			Address: "host_address",
		},
		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Address:      "trade_address",
			ConnectionId: ibctesting.FirstConnectionID,
			Type:         types.ICAAccountType_CONVERTER_TRADE,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)

	// Create epoch tracker to dictate timeout
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, time.Second*10)

	// Confirm the sequence number was incremented after a successful send
	startSequence := s.MustGetNextSequenceNumber(portId, channelId)

	err := s.App.StakeibcKeeper.TransferConvertedTokensTradeToHost(s.Ctx, transferAmount, route)
	s.Require().NoError(err, "no error expected when transfering tokens")

	endSequence := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(startSequence+1, endSequence, "sequence number should have incremented from transfer")

	// Attempt to send without a valid ICA address - it should fail
	invalidRoute := route
	invalidRoute.HostAccount.Address = ""
	err = s.App.StakeibcKeeper.TransferConvertedTokensTradeToHost(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no host account found")

	invalidRoute = route
	invalidRoute.TradeAccount.Address = ""
	err = s.App.StakeibcKeeper.TransferConvertedTokensTradeToHost(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no trade account found")
}

// --------------------------------------------------------------
//            Trade Route ICQ Test Helpers
// --------------------------------------------------------------

// Helper function to validate the address and denom from the query request data
func (s *KeeperTestSuite) validateAddressAndDenomInRequest(data []byte, expectedAddress, expectedDenom string) {
	actualAddress, actualDenom := s.ExtractAddressAndDenomFromBankPrefix(data)
	s.Require().Equal(expectedAddress, actualAddress, "query account address")
	s.Require().Equal(expectedDenom, actualDenom, "query denom")
}

// Helper function to validate the trade route query callback data
func (s *KeeperTestSuite) validateTradeRouteQueryCallback(actualCallbackDataBz []byte) {
	expectedCallbackData := types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	}

	var actualCallbackData types.TradeRouteCallback
	err := proto.Unmarshal(actualCallbackDataBz, &actualCallbackData)
	s.Require().NoError(err)
	s.Require().Equal(expectedCallbackData, actualCallbackData, "query callback data")
}

// --------------------------------------------------------------
//            Withdrawal Account - Reward Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the withdrawal address and the
// reward_denom_on_host since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupWithdrawalRewardBalanceQueryTestCase() (route types.TradeRoute, expectedTimeout time.Duration) {
	// Create a transfer channel so the connection exists for the query submission
	s.CreateTransferChannel(HostChainId)

	// Create and set the trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		RewardDenomOnHostZone:   "ibc/reward_on_host",
		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
			Address:      StrideICAAddress, // must be a valid bech32, easiest to use stride prefix for validation
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create and set the epoch tracker for timeouts (the timeout is halfway through the epoch)
	epochDuration := time.Second * 30
	expectedTimeout = epochDuration / 2
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, epochDuration)

	return tradeRoute, expectedTimeout
}

// Tests a successful WithdrawalRewardBalanceQuery
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Successful() {
	route, timeoutDuration := s.SetupWithdrawalRewardBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, route)
	s.Require().NoError(err, "no error expected when querying balance")

	// Validate fields from ICQ submission
	expectedRequestData := s.GetBankStoreKeyPrefix(StrideICAAddress, route.RewardDenomOnHostZone)

	query := s.ValidateQuerySubmission(
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		expectedRequestData,
		keeper.ICQCallbackID_WithdrawalRewardBalance,
		timeoutDuration,
		icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	)

	s.validateAddressAndDenomInRequest(query.RequestData, route.HostAccount.Address, route.RewardDenomOnHostZone)
	s.validateTradeRouteQueryCallback(query.CallbackData)
}

// Tests a WithdrawalRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute, _ := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Change the withdrawal ICA account address to be invalid
	tradeRoute.HostAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid withdrawal account address")
}

// Tests a WithdrawalRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_MissingEpoch() {
	tradeRoute, _ := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a WithdrawalRewardBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_FailedQuerySubmission() {
	tradeRoute, _ := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Change the withdrawal ICA connection id to be invalid
	tradeRoute.HostAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//            Trade Account - Converted Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
// host_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeConvertedBalanceQueryTestCase() (route types.TradeRoute, expectedTimeout time.Duration) {
	// Create a transfer channel so the connection exists for the query submission
	s.CreateTransferChannel(HostChainId)

	// Create and set the trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		HostDenomOnTradeZone:    "ibc/host_on_trade",
		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
			Address:      StrideICAAddress, // must be a valid bech32, easiest to use stride prefix for validation
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create and set the epoch tracker for timeouts
	timeoutDuration := time.Second * 30
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, timeoutDuration)

	return tradeRoute, timeoutDuration
}

// Tests a successful TradeConvertedBalanceQuery
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Successful() {
	route, timeoutDuration := s.SetupTradeConvertedBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, route)
	s.Require().NoError(err, "no error expected when querying balance")

	// Validate fields from ICQ submission
	expectedRequestData := s.GetBankStoreKeyPrefix(StrideICAAddress, route.HostDenomOnTradeZone)

	query := s.ValidateQuerySubmission(
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		expectedRequestData,
		keeper.ICQCallbackID_TradeConvertedBalance,
		timeoutDuration,
		icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	)

	s.validateAddressAndDenomInRequest(query.RequestData, route.TradeAccount.Address, route.HostDenomOnTradeZone)
	s.validateTradeRouteQueryCallback(query.CallbackData)
}

// Tests a TradeConvertedBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute, _ := s.SetupTradeConvertedBalanceQueryTestCase()

	// Change the trade ICA account address to be invalid
	tradeRoute.TradeAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}

// Tests a TradeConvertedBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_MissingEpoch() {
	tradeRoute, _ := s.SetupTradeConvertedBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a TradeConvertedBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_FailedQuerySubmission() {
	tradeRoute, _ := s.SetupTradeConvertedBalanceQueryTestCase()

	// Change the trade ICA connection id to be invalid
	tradeRoute.TradeAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}
