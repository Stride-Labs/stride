package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const chainId = "GAIA"

type TransferCommunityPoolTokensTestCase struct {
	hostZone	types.HostZone
	coin 		sdk.Coin
	action 		string
}

func (s *KeeperTestSuite) SetupTransferCommunityPoolTokens() TransferCommunityPoolTokensTestCase {
	s.CreateICAChannel(chainId+"."+types.ICAAccountType_COMMUNITY_POOL_DEPOSIT.String())

	holdingAccount := s.TestAccs[0]
	holdingAddress := holdingAccount.String()
	depositIcaAccount := s.TestAccs[1]
	depositIcaAddress := depositIcaAccount.String()
	hostZone := types.HostZone{
		ChainId: chainId,
		ConnectionId: "connection-0",
		TransferChannelId: "channel-0",
		CommunityPoolHoldingAddress: holdingAddress,
		CommunityPoolDepositIcaAddress: depositIcaAddress,
	}

	balanceToTransfer := sdkmath.NewInt(1_000_000)
	coin := sdk.NewCoin("tokens", balanceToTransfer)
	s.FundAccount(depositIcaAccount, coin)

	return TransferCommunityPoolTokensTestCase{
		hostZone: hostZone,
		coin: coin,
		action: keeper.LiquidStake,
	}
}

func (s *KeeperTestSuite) TestTransferCommunityPoolTokens_MissingHoldingFail() {
	tc := s.SetupTransferCommunityPoolTokens()
	tc.hostZone.CommunityPoolHoldingAddress = ""

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolTokens(s.Ctx, tc.hostZone, tc.coin, tc.action)
	s.Require().ErrorContains(err, "holding address")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolTokens_MissingDepositFail() {
	tc := s.SetupTransferCommunityPoolTokens()
	tc.hostZone.CommunityPoolDepositIcaAddress = ""

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolTokens(s.Ctx, tc.hostZone, tc.coin, tc.action)
	s.Require().ErrorContains(err, "deposit address")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolTokens_ConnectionSendFail() {
	tc := s.SetupTransferCommunityPoolTokens()
	tc.hostZone.ConnectionId = "MissingChannel"

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolTokens(s.Ctx, tc.hostZone, tc.coin, tc.action)
	s.Require().ErrorContains(err, "invalid connection id")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolTokens_Successful() {
	tc := s.SetupTransferCommunityPoolTokens()

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolTokens(s.Ctx, tc.hostZone, tc.coin, tc.action)
	s.Require().NoError(err)

	// Get the tx packets which are being sent
	// Get the marshalled msg data --> verify the fields are as expected in transfer message
	//   verify the autopilot message is correct
}




type TransferCoinToReturnTestCase struct {
	hostZone 	types.HostZone
	coin   		sdk.Coin
}

func (s *KeeperTestSuite) SetupTransferCoinToReturn() TransferCoinToReturnTestCase {
	s.CreateTransferChannel(chainId)	

	holdingAccount := s.TestAccs[0]
	holdingAddress := holdingAccount.String()
	returnIcaAddress := s.TestAccs[1].String()
	hostZone := types.HostZone{
		ChainId: chainId,
		TransferChannelId: "channel-0",
		CommunityPoolHoldingAddress: holdingAddress,
		CommunityPoolReturnIcaAddress: returnIcaAddress,
	}

	balanceToTransfer := sdkmath.NewInt(1_000_000)
	coin := sdk.NewCoin("tokens", balanceToTransfer)
	s.FundAccount(holdingAccount, coin)

	return TransferCoinToReturnTestCase{
		hostZone: hostZone,
		coin: coin,
	}
}

func (s *KeeperTestSuite) TestTransferCoinToReturn_ChannelTransferFail() {
	tc := s.SetupTransferCoinToReturn()
	tc.hostZone.TransferChannelId = "WrongChannel"

	// Verify that the transfer was successfully sent off
	err := s.App.StakeibcKeeper.TransferCoinToReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "Error submitting ibc transfer")
}

func (s *KeeperTestSuite) TestTransferCoinToReturn_MissingTokens() {
	tc := s.SetupTransferCoinToReturn()
	tc.coin.Denom = "MissingDenom"

	// Verify that the transfer was successfully sent off
	err := s.App.StakeibcKeeper.TransferCoinToReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "Error submitting ibc transfer")
	s.Require().ErrorContains(err, "insufficient funds")
}

func (s *KeeperTestSuite) TestTransferCoinToReturn_Successful() {
	tc := s.SetupTransferCoinToReturn()

	// Verify that the transfer was successfully sent off
	err := s.App.StakeibcKeeper.TransferCoinToReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().NoError(err)	

	// Verify that the ICA callback was manually set
	callbacks := s.App.StakeibcKeeper.ICACallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Equal(1, len(callbacks), "there should be one ica callback")
	callback := callbacks[0]
	s.Require().Equal(keeper.ICACallbackID_CommunityPoolReturn, callback.CallbackId, 
		"the registered callback function id should match")

	// Verify that the callbackArgs has a legal correct type which can unmarshal
	callbackArgs := types.CommunityPoolReturnTransferCallback{}
	err = proto.Unmarshal(callback.CallbackArgs, &callbackArgs)
	s.Require().NoError(err)

	// Check that the callback was built with expected args
	s.Require().Equal(tc.hostZone.ChainId, callbackArgs.HostZoneId, "the chainId should match")
	s.Require().Equal(tc.coin.Amount, callbackArgs.Amount, "amount saved in callback should match")
	s.Require().Equal(tc.coin.Denom, callbackArgs.DenomStride, "denom saved in callback should match")
	expectedIbcDenom := s.App.StakeibcKeeper.GetDenomOnHostZone(tc.coin.Denom, tc.hostZone)
	s.Require().Equal(expectedIbcDenom, callbackArgs.IbcDenom, "ibc denom in callback should match")		
}

func (s *KeeperTestSuite) TestTransferCoinToReturn_Sequence() {
	tc := s.SetupTransferCoinToReturn()
	transferPort := "transfer"
	sequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx,
		transferPort, tc.hostZone.TransferChannelId)
	s.Require().True(found)

	err := s.App.StakeibcKeeper.TransferCoinToReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().NoError(err)

	// Verify that the ICA callback was manually set and has the expected sequence
	callbacks := s.App.StakeibcKeeper.ICACallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Equal(1, len(callbacks), "there should be one ica callback")
	callback := callbacks[0]
	s.Require().Equal(sequence, callback.Sequence, "transfer msg sequence should be equal")
}


type GetDenomOnHostZoneTestCase struct {
	strideDenom 	string
	transferChannel	string
	ibcDenom		string
}

func (s *KeeperTestSuite) TestGetDenomOnHostZone() {
	testCases := []GetDenomOnHostZoneTestCase{
		{
			strideDenom: "uatom", 
			transferChannel: "channel-0",
			ibcDenom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		},
		{
			strideDenom: "uatom", 
			transferChannel: "channel-3",
			ibcDenom: "ibc/A4DB47A9D3CF9A068D454513891B526702455D3EF08FB9EB558C561F9DC2B701",
		},		
	}

	hostZone := types.HostZone{
		TransferChannelId: "",
	}
	for _, tc := range testCases {
		hostZone.TransferChannelId = tc.transferChannel
		computedDenom := s.App.StakeibcKeeper.GetDenomOnHostZone(tc.strideDenom, hostZone)
		s.Require().Equal(computedDenom, tc.ibcDenom, "ibcDenom should match known value")
	}	
}
