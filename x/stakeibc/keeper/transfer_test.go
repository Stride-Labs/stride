package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
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
		action:    keeper.LiquidStake,
		channelId: channelId,
		portId:    portId,
	}
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_Successful() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.portId, tc.channelId)
	s.Require().True(found)

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin, tc.action)
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
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin, tc.action)
	s.Require().ErrorContains(err, "holding address")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_MissingDepositFail() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()
	tc.hostZone.CommunityPoolDepositIcaAddress = ""

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin, tc.action)
	s.Require().ErrorContains(err, "deposit address")
}

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_ConnectionSendFail() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()
	tc.hostZone.ConnectionId = "MissingChannel"

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin, tc.action)
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

	// Verify that the ICA callback was manually set
	callbacks := s.App.StakeibcKeeper.ICACallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Equal(1, len(callbacks), "there should be one ica callback")
	callback := callbacks[0]
	s.Require().Equal(startSequence, callback.Sequence, "transfer msg sequence should be equal")
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

type GetDenomOnHostZoneTestCase struct {
	strideDenom     string
	transferChannel string
	ibcDenom        string
}

func (s *KeeperTestSuite) TestGetDenomOnHostZone() {
	testCases := []GetDenomOnHostZoneTestCase{
		{
			strideDenom:     "uatom",
			transferChannel: "channel-0",
			ibcDenom:        "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		},
		{
			strideDenom:     "uatom",
			transferChannel: "channel-3",
			ibcDenom:        "ibc/A4DB47A9D3CF9A068D454513891B526702455D3EF08FB9EB558C561F9DC2B701",
		},
		{
			strideDenom:     "uatom",
			transferChannel: "channel-10",
			ibcDenom:        "ibc/A670D9568B3E399316EEDE40C1181B7AA4BD0695F0B37513CE9B95B977DFC12E",
		},
		{
			strideDenom:     "uosmo",
			transferChannel: "channel-999",
			ibcDenom:        "ibc/BBF0BA1A51EA726A21CDC784B4834DCB64407BB6E2BFC8F15DE266DB05F6000D",
		},
		{
			strideDenom:     "uusdc",
			transferChannel: "channel-208",
			ibcDenom:        "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
		},
		{
			strideDenom:     "ujuno",
			transferChannel: "channel-42",
			ibcDenom:        "ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED",
		},
		{
			strideDenom:     "aevmos",
			transferChannel: "channel-204",
			ibcDenom:        "ibc/6AE98883D4D5D5FF9E50D7130F1305DA2FFA0C652D1DD9C123657C6B4EB2DF8A",
		},
		{
			strideDenom:     "ustrd",
			transferChannel: "channel-326",
			ibcDenom:        "ibc/A8CA5EE328FA10C9519DF6057DA1F69682D28F7D0F5CCC7ECB72E3DCA2D157A4",
		},
		{
			strideDenom:     "stuatom",
			transferChannel: "channel-326",
			ibcDenom:        "ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901",
		},
		{
			strideDenom:     "stuosmo",
			transferChannel: "channel-326",
			ibcDenom:        "ibc/D176154B0C63D1F9C6DCFB4F70349EBF2E2B5A87A05902F57A6AE92B863E9AEC",
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
