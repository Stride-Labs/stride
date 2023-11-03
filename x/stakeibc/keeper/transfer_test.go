package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const chainId = "GAIA"

type TransferCommunityPoolDepositToHoldingTestCase struct {
	hostZone  types.HostZone
	coin      sdk.Coin
	action    string
	channelId string
	portId    string
}

func (s *KeeperTestSuite) SetupTransferCommunityPoolDepositToHolding() TransferCommunityPoolDepositToHoldingTestCase {
	owner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	channelId, portId := s.CreateICAChannel(owner)

	holdingAccount := s.TestAccs[0]
	holdingAddress := holdingAccount.String()
	depositIcaAccount := s.TestAccs[1]
	depositIcaAddress := depositIcaAccount.String()
	hostZone := types.HostZone{
		ChainId:                        chainId,
		ConnectionId:                   "connection-0",
		TransferChannelId:              "channel-0",
		CommunityPoolHoldingAddress:    holdingAddress,
		CommunityPoolDepositIcaAddress: depositIcaAddress,
	}

	balanceToTransfer := sdkmath.NewInt(1_000_000)
	coin := sdk.NewCoin("tokens", balanceToTransfer)
	s.FundAccount(depositIcaAccount, coin)

	return TransferCommunityPoolDepositToHoldingTestCase{
		hostZone:  hostZone,
		coin:      coin,
		channelId: channelId,
		portId:    portId,
	}
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_Successful() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.portId, tc.channelId)
	s.Require().True(found)

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin)
	s.Require().NoError(err)

	// Verify the ICA sequence number incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.portId, tc.channelId)
	s.Require().True(found)
	s.Require().Equal(endSequence, startSequence+1, "sequence number should have incremented")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_MissingHoldingFail() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()
	tc.hostZone.CommunityPoolHoldingAddress = ""

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "holding address")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_MissingDepositFail() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()
	tc.hostZone.CommunityPoolDepositIcaAddress = ""

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "deposit address")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_ConnectionSendFail() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()
	tc.hostZone.ConnectionId = "MissingChannel"

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "invalid connection id")
}

type TransferHoldingToCommunityPoolReturnTestCase struct {
	hostZone types.HostZone
	coin     sdk.Coin
}

func (s *KeeperTestSuite) TestTransferHoldingToCommunityPoolReturn_Successful() {
	tc := s.SetupTransferHoldingToCommunityPoolReturn()

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx,
		transfertypes.PortID, tc.hostZone.TransferChannelId)
	s.Require().True(found)

	// Verify that the transfer was successfully sent off
	err := s.App.StakeibcKeeper.TransferHoldingToCommunityPoolReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().NoError(err)

	// Verify the transfer sequence number incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx,
		transfertypes.PortID, tc.hostZone.TransferChannelId)
	s.Require().True(found)
	s.Require().Equal(endSequence, startSequence+1, "sequence number should have incremented")
}

func (s *KeeperTestSuite) SetupTransferHoldingToCommunityPoolReturn() TransferHoldingToCommunityPoolReturnTestCase {
	s.CreateTransferChannel(chainId)

	holdingAccount := s.TestAccs[0]
	holdingAddress := holdingAccount.String()
	returnIcaAddress := s.TestAccs[1].String()
	hostZone := types.HostZone{
		ChainId:                       chainId,
		TransferChannelId:             "channel-0",
		CommunityPoolHoldingAddress:   holdingAddress,
		CommunityPoolReturnIcaAddress: returnIcaAddress,
	}

	balanceToTransfer := sdkmath.NewInt(1_000_000)
	coin := sdk.NewCoin("tokens", balanceToTransfer)
	s.FundAccount(holdingAccount, coin)

	return TransferHoldingToCommunityPoolReturnTestCase{
		hostZone: hostZone,
		coin:     coin,
	}
}

func (s *KeeperTestSuite) TestTransferHoldingToCommunityPoolReturn_ChannelTransferFail() {
	tc := s.SetupTransferHoldingToCommunityPoolReturn()
	tc.hostZone.TransferChannelId = "WrongChannel"

	// Verify that the transfer was successfully sent off
	err := s.App.StakeibcKeeper.TransferHoldingToCommunityPoolReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "Error submitting ibc transfer")
}

func (s *KeeperTestSuite) TestTransferHoldingToCommunityPoolReturn_MissingTokens() {
	tc := s.SetupTransferHoldingToCommunityPoolReturn()
	tc.coin.Denom = "MissingDenom"

	// Verify that the transfer was successfully sent off
	err := s.App.StakeibcKeeper.TransferHoldingToCommunityPoolReturn(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "Error submitting ibc transfer")
	s.Require().ErrorContains(err, "insufficient funds")
}

