package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	sdkmath "cosmossdk.io/math"

	epochtypes "github.com/Stride-Labs/stride/v30/x/epochs/types"
	"github.com/Stride-Labs/stride/v30/x/stakeibc/types"
)

const chainId = "GAIA"

type TransferCommunityPoolDepositToHoldingTestCase struct {
	hostZone  types.HostZone
	coin      sdk.Coin
	channelId string
	portId    string
}

func (s *KeeperTestSuite) SetupTransferCommunityPoolDepositToHolding() TransferCommunityPoolDepositToHoldingTestCase {
	owner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	channelId, portId := s.CreateICAChannel(owner)

	holdingAddress := s.TestAccs[0].String()
	depositIcaAccount := s.TestAccs[1]
	depositIcaAddress := depositIcaAccount.String()
	hostZone := types.HostZone{
		ChainId:                          chainId,
		ConnectionId:                     ibctesting.FirstConnectionID,
		TransferChannelId:                ibctesting.FirstChannelID,
		HostDenom:                        Atom,
		CommunityPoolStakeHoldingAddress: holdingAddress,
		CommunityPoolDepositIcaAddress:   depositIcaAddress,
	}

	strideEpoch := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // used for transfer timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpoch)

	balanceToTransfer := sdkmath.NewInt(1_000_000)
	coin := sdk.NewCoin(Atom, balanceToTransfer)
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

func (s *KeeperTestSuite) TestTransferCommunityPoolDepositToHolding_MissingStakeAddressFail() {
	tc := s.SetupTransferCommunityPoolDepositToHolding()
	tc.hostZone.CommunityPoolStakeHoldingAddress = ""

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
	tc.hostZone.ConnectionId = "MissingConnection"

	// Verify that the ICA msg was successfully sent off
	err := s.App.StakeibcKeeper.TransferCommunityPoolDepositToHolding(s.Ctx, tc.hostZone, tc.coin)
	s.Require().ErrorContains(err, "connection MissingConnection not found")
}

type TransferHoldingToCommunityPoolReturnTestCase struct {
	hostZone types.HostZone
	coin     sdk.Coin
}

func (s *KeeperTestSuite) SetupTransferHoldingToCommunityPoolReturn() TransferHoldingToCommunityPoolReturnTestCase {
	s.CreateTransferChannel(chainId)

	holdingAccount := s.TestAccs[0]
	holdingAddress := holdingAccount.String()
	returnIcaAddress := s.TestAccs[1].String()
	hostZone := types.HostZone{
		ChainId:                          chainId,
		TransferChannelId:                ibctesting.FirstChannelID,
		HostDenom:                        Atom,
		CommunityPoolStakeHoldingAddress: holdingAddress,
		CommunityPoolReturnIcaAddress:    returnIcaAddress,
	}

	strideEpoch := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // used for transfer timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpoch)

	balanceToTransfer := sdkmath.NewInt(1_000_000)
	coin := sdk.NewCoin(Atom, balanceToTransfer)
	s.FundAccount(holdingAccount, coin)

	return TransferHoldingToCommunityPoolReturnTestCase{
		hostZone: hostZone,
		coin:     coin,
	}
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

func (s *KeeperTestSuite) TestGetStIbcDenomOnHostZone() {
	testCases := []struct {
		hostDenom        string
		channelOnStride  string
		channelOnHost    string
		expectedIBCDenom string
	}{
		{
			hostDenom:        "uatom",
			channelOnStride:  "channel-0",
			channelOnHost:    "channel-391",
			expectedIBCDenom: "ibc/B05539B66B72E2739B986B86391E5D08F12B8D5D2C2A7F8F8CF9ADF674DFA231",
		},
		{
			hostDenom:        "uosmo",
			channelOnStride:  "channel-5",
			channelOnHost:    "channel-326",
			expectedIBCDenom: "ibc/D176154B0C63D1F9C6DCFB4F70349EBF2E2B5A87A05902F57A6AE92B863E9AEC",
		},
		{
			hostDenom:        "ujuno",
			channelOnStride:  "channel-24",
			channelOnHost:    "channel-139",
			expectedIBCDenom: "ibc/F4F5F27F40F927F8A4FF9F5601F80AD5D77B366570E7C59856B8CE4135AC1F59",
		},
		{
			hostDenom:        "ustars",
			channelOnStride:  "channel-19",
			channelOnHost:    "channel-106",
			expectedIBCDenom: "ibc/7A58490427EF0092E2BFFB4BEEBA38E29B09E9B98557DFC78335B43F15CF2676",
		},
		{
			hostDenom:        "inj",
			channelOnStride:  "channel-6",
			channelOnHost:    "channel-89",
			expectedIBCDenom: "ibc/AC87717EA002B0123B10A05063E69BCA274BA2C44D842AEEB41558D2856DCE93",
		},
	}

	// Create each channel on stride with the associated host channel as a counterparty
	for _, tc := range testCases {
		channel := channeltypes.Channel{
			Counterparty: channeltypes.Counterparty{
				ChannelId: tc.channelOnHost,
			},
		}
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, tc.channelOnStride, channel)
	}

	// For each case, check the generated IBC denom
	for _, tc := range testCases {
		hostZone := types.HostZone{
			TransferChannelId: tc.channelOnStride,
			HostDenom:         tc.hostDenom,
		}
		actualIBCDenom, err := s.App.StakeibcKeeper.GetStIbcDenomOnHostZone(s.Ctx, hostZone)
		s.Require().NoError(err, "no error expected when generating IBC denom")
		s.Require().Equal(tc.expectedIBCDenom, actualIBCDenom, "stToken ibc denom")
	}

	// Test a non-existent channel ID
	invalidHostZone := types.HostZone{TransferChannelId: "channel-1000"}
	_, err := s.App.StakeibcKeeper.GetStIbcDenomOnHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "channel not found")

	// Test a channel that has a non-transfer port
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, "different port", "channel-1000", channeltypes.Channel{})
	_, err = s.App.StakeibcKeeper.GetStIbcDenomOnHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "channel not found")

	// Test a with an empty counterparty channel
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, "channel-1000", channeltypes.Channel{})
	_, err = s.App.StakeibcKeeper.GetStIbcDenomOnHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "counterparty channel not found")
}
