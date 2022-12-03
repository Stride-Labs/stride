package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type WithdrawalBalanceICQCallbackState struct {
	hostZone          stakeibctypes.HostZone
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
	coin := sdk.NewCoin(denom, sdk.NewInt(amount))
	coinBz := s.App.RecordsKeeper.Cdc.MustMarshal(&coin)
	return coinBz
}

func (s *KeeperTestSuite) SetupWithdrawalBalanceCallbackTest() WithdrawalBalanceICQCallbackTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	withdrawalChannelId := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]

	feeAddress := "cosmos_FEE"

	hostZone := stakeibctypes.HostZone{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		DelegationAccount: &stakeibctypes.ICAAccount{
			Address: delegationAddress,
			Target:  stakeibctypes.ICAAccountType_DELEGATION,
		},
		WithdrawalAccount: &stakeibctypes.ICAAccount{
			Address: withdrawalAddress,
			Target:  stakeibctypes.ICAAccountType_WITHDRAWAL,
		},
		FeeAccount: &stakeibctypes.ICAAccount{
			Address: feeAddress,
			Target:  stakeibctypes.ICAAccountType_FEE,
		},
	}

	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	withdrawalBalance := int64(1000)
	commission := uint64(10)
	expectedReinvestment := sdk.NewCoin(Atom, sdk.NewInt(int64(900)))

	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	params.StrideCommission = uint64(commission)

	queryResponse := s.CreateBalanceQueryResponse(withdrawalBalance, Atom)

	return WithdrawalBalanceICQCallbackTestCase{
		initialState: WithdrawalBalanceICQCallbackState{
			hostZone: hostZone,
			withdrawalChannel: Channel{
				PortID:    icatypes.PortPrefix + withdrawalAccountOwner,
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

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_Successful() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	withdrawalChannel := tc.initialState.withdrawalChannel
	withdrawalPortId := withdrawalChannel.PortID
	withdrawalChannelId := withdrawalChannel.ChannelID

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, withdrawalPortId, withdrawalChannelId)
	s.Require().True(found, "sequence number not found before reinvestment")

	// Call the ICQ callback
	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
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

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_EmptyCallbackArgs() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Replace the query response an empty byte array (this happens when the account has not been registered yet)
	emptyCallbackArgs := []byte{}

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, emptyCallbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_ZeroBalance() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Replace the query response with a coin that has a nil amount
	tc.validArgs.callbackArgs = s.CreateBalanceQueryResponse(0, Atom)

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_ZeroBalanceImplied() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Replace the query response with a coin that has a nil amount
	coin := sdk.Coin{}
	coinBz := s.App.RecordsKeeper.Cdc.MustMarshal(&coin)
	tc.validArgs.callbackArgs = coinBz

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_HostZoneNotFound() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Submit callback with incorrect host zone
	invalidQuery := tc.validArgs.query
	invalidQuery.ChainId = "fake_host_zone"
	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_InvalidArgs() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")
	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)

	expectedErrMsg := "unable to unmarshal balance in callback args for zone: GAIA, "
	expectedErrMsg += "err: unexpected EOF: unable to marshal data structure"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_NoWithdrawalAccount() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Remove the withdrawal account
	badHostZone := tc.initialState.hostZone
	badHostZone.WithdrawalAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	expectedErrMsg := "WithdrawalBalanceCallback: no withdrawal account found for zone: GAIA: "
	expectedErrMsg += "ICA acccount not found on host zone"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_NoDelegationAccount() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Remove the delegation account
	badHostZone := tc.initialState.hostZone
	badHostZone.DelegationAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	expectedErrMsg := "WithdrawalBalanceCallback: no delegation account found for zone: GAIA: "
	expectedErrMsg += "ICA acccount not found on host zone"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_NoFeeAccount() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Remove the fee account
	badHostZone := tc.initialState.hostZone
	badHostZone.FeeAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	expectedErrMsg := "WithdrawalBalanceCallback: no fee account found for zone: GAIA: "
	expectedErrMsg += "ICA acccount not found on host zone"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_FailedSubmitTx() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	// Remove connectionId from host zone so the ICA tx fails
	badHostZone := tc.initialState.hostZone
	badHostZone.ConnectionId = "connection-X"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "Failed to SubmitTxs for GAIA - connection-X")
	s.Require().ErrorContains(err, "invalid connection id, connection-X not found")
}
