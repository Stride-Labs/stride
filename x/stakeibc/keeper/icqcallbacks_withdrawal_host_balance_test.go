package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v27/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	icqtypes "github.com/Stride-Labs/stride/v27/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

type WithdrawalBalanceICQCallbackState struct {
	hostZone          types.HostZone
	withdrawalChannel Channel
	withdrawalBalance int64
}

type WithdrawalBalanceICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type WithdrawalBalanceICQCallbackTestCase struct {
	initialState         WithdrawalBalanceICQCallbackState
	validArgs            WithdrawalBalanceICQCallbackArgs
	expectedReinvestment sdk.Coin
}

// The response from the WithdrawalBalance ICQ is a serialized sdk.Coin containing
// the address' balance. This function creates the serialized response
func (s *KeeperTestSuite) CreateBalanceQueryResponse(amount int64, denom string) []byte {
	coin := sdk.NewCoin(denom, sdkmath.NewInt(amount))
	coinBz := s.App.AppCodec().MustMarshal(&coin)
	return coinBz
}

func (s *KeeperTestSuite) SetupWithdrawalHostBalanceCallbackTest() WithdrawalBalanceICQCallbackTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	withdrawalChannelId, withdrawalPortId := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]

	feeAddress := "cosmos_FEE"

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		ConnectionId:         ibctesting.FirstConnectionID,
		DelegationIcaAddress: delegationAddress,
		WithdrawalIcaAddress: withdrawalAddress,
		FeeIcaAddress:        feeAddress,
	}

	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	withdrawalBalance := int64(1000)
	expectedReinvestment := sdk.NewCoin(Atom, sdkmath.NewInt(int64(900)))

	queryResponse := s.CreateBalanceQueryResponse(withdrawalBalance, Atom)

	return WithdrawalBalanceICQCallbackTestCase{
		initialState: WithdrawalBalanceICQCallbackState{
			hostZone: hostZone,
			withdrawalChannel: Channel{
				PortID:    withdrawalPortId,
				ChannelID: withdrawalChannelId,
			},
			withdrawalBalance: withdrawalBalance,
		},
		validArgs: WithdrawalBalanceICQCallbackArgs{
			query: icqtypes.Query{
				Id:      "0",
				ChainId: HostChainId,
			},
			callbackArgs: queryResponse,
		},
		expectedReinvestment: expectedReinvestment,
	}
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_Successful() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	withdrawalChannel := tc.initialState.withdrawalChannel
	withdrawalPortId := withdrawalChannel.PortID
	withdrawalChannelId := withdrawalChannel.ChannelID

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, withdrawalPortId, withdrawalChannelId)
	s.Require().True(found, "sequence number not found before reinvestment")

	// Call the ICQ callback
	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm ICA reinvestment callback data was stored
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 1, "number of callbacks found")
	callbackKey := icacallbackstypes.PacketID(withdrawalPortId, withdrawalChannelId, startSequence)
	callbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
	s.Require().True(found, "callback data was not found for callback key (%s)", callbackKey)
	s.Require().Equal("reinvest", callbackData.CallbackId, "callback ID")

	// Confirm reinvestment callback args
	callbackArgs, err := s.App.StakeibcKeeper.UnmarshalReinvestCallbackArgs(s.Ctx, callbackData.CallbackArgs)
	s.Require().NoError(err, "unmarshalling callback args error for callback key (%s)", callbackKey)
	s.Require().Equal(tc.initialState.hostZone.ChainId, callbackArgs.HostZoneId, "host zone in callback args (%s)", callbackKey)
	s.Require().Equal(tc.expectedReinvestment, callbackArgs.ReinvestAmount, "reinvestment coin in callback args (%s)", callbackKey)

	// Confirm the sequence number was incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, withdrawalPortId, withdrawalChannelId)
	s.Require().True(found, "sequence number not found after reinvestment")
	s.Require().Equal(endSequence, startSequence+1, "sequence number after reinvestment")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_EmptyCallbackArgs() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Replace the query response an empty byte array (this happens when the account has not been registered yet)
	emptyCallbackArgs := []byte{}

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, emptyCallbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_ZeroBalance() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Replace the query response with a coin that has a nil amount
	tc.validArgs.callbackArgs = s.CreateBalanceQueryResponse(0, Atom)

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_ZeroBalanceImplied() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Replace the query response with a coin that has a nil amount
	coin := sdk.Coin{}
	coinBz := s.App.RecordsKeeper.Cdc.MustMarshal(&coin)
	tc.validArgs.callbackArgs = coinBz

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_HostZoneNotFound() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Submit callback with incorrect host zone
	invalidQuery := tc.validArgs.query
	invalidQuery.ChainId = "fake_host_zone"
	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_InvalidArgs() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")
	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)

	s.Require().ErrorContains(err, "unable to determine balance from query response")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_NoWithdrawalAccount() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Remove the withdrawal account
	badHostZone := tc.initialState.hostZone
	badHostZone.WithdrawalIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no withdrawal account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_NoDelegationAccount() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Remove the delegation account
	badHostZone := tc.initialState.hostZone
	badHostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no delegation account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_NoFeeAccount() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Remove the fee account
	badHostZone := tc.initialState.hostZone
	badHostZone.FeeIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no fee account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_FailedToCheckForRebate() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Add a rebate to the host zone - since there are no stTokens in supply, the test will fail
	hostZone := s.MustGetHostZone(HostChainId)
	hostZone.CommunityPoolRebate = &types.CommunityPoolRebate{
		RebateRate:                sdk.MustNewDecFromStr("0.5"),
		LiquidStakedStTokenAmount: sdkmath.NewInt(1),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "unable to split reward amount into fee and reinvest amounts")
}

func (s *KeeperTestSuite) TestWithdrawalHostBalanceCallback_FailedSubmitTx() {
	tc := s.SetupWithdrawalHostBalanceCallbackTest()

	// Remove connectionId from host zone so the ICA tx fails
	badHostZone := tc.initialState.hostZone
	badHostZone.ConnectionId = "connection-X"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := keeper.WithdrawalHostBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "Failed to SubmitTxs")
	s.Require().ErrorContains(err, "connection connection-X not found")
}
