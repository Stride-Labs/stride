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
