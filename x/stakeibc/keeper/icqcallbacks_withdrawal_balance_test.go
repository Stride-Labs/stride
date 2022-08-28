package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type WithdrawalBalanceCallbackState struct {
	hostZone          stakeibctypes.HostZone
	withdrawalChannel Channel
	withdrawalBalance int64
}

type WithdrawalBalanceCallbackArgs struct {
	query    icqtypes.Query
	response []byte
}

type WithdrawalBalanceCallbackTestCase struct {
	initialState         WithdrawalBalanceCallbackState
	validArgs            WithdrawalBalanceCallbackArgs
	expectedReinvestment sdk.Coin
}

func (s *KeeperTestSuite) CreateBalanceQueryRequest(address string, denom string) []byte {
	_, addressBz, err := bech32.DecodeAndConvert(address)
	s.Require().NoError(err)

	denomBz := []byte(denom)
	balancePrefix := banktypes.CreateAccountBalancesPrefix(addressBz)
	return append(balancePrefix, denomBz...)
}

func (s *KeeperTestSuite) CreateBalanceQueryResponse(amount int64, denom string) []byte {
	coin := sdk.NewCoin(denom, sdk.NewInt(amount))
	coinBz := s.App.RecordsKeeper.Cdc.MustMarshal(&coin)
	return coinBz
}

func (s *KeeperTestSuite) SetupWithdrawalBalanceCallbackTest() WithdrawalBalanceCallbackTestCase {
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

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), strideEpochTracker)

	withdrawalBalance := int64(1000)
	commission := uint64(10)
	expectedReinvestment := sdk.NewCoin(Atom, sdk.NewInt(int64(900)))

	params := s.App.StakeibcKeeper.GetParams(s.Ctx())
	params.StrideCommission = uint64(commission)

	queryRequest := s.CreateBalanceQueryRequest(withdrawalAddress, Atom)
	queryResponse := s.CreateBalanceQueryResponse(withdrawalBalance, Atom)

	return WithdrawalBalanceCallbackTestCase{
		initialState: WithdrawalBalanceCallbackState{
			hostZone: hostZone,
			withdrawalChannel: Channel{
				PortID:    "icacontroller-" + withdrawalAccountOwner,
				ChannelID: withdrawalChannelId,
			},
			withdrawalBalance: withdrawalBalance,
		},
		validArgs: WithdrawalBalanceCallbackArgs{
			query: icqtypes.Query{
				Id:      "0",
				ChainId: HostChainId,
				Request: queryRequest,
			},
			response: queryResponse,
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

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), withdrawalPortId, withdrawalChannelId)
	s.Require().True(found, "sequence number not found before reinvestment")

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx(), tc.validArgs.response, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm callback data was stored
	callbackKey := icacallbackstypes.PacketID(withdrawalPortId, withdrawalChannelId, startSequence)
	callbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx(), callbackKey)
	s.Require().True(found, "callback data was not found for callback key (%s)", callbackKey)
	s.Require().Equal("reinvest", callbackData.CallbackId, "callback ID")

	// Confirm callback args
	callbackArgs, err := s.App.StakeibcKeeper.UnmarshalReinvestCallbackArgs(s.Ctx(), callbackData.CallbackArgs)
	s.Require().NoError(err, "unmarshalling callback args error for callback key (%s)", callbackKey)
	s.Require().Equal(tc.initialState.hostZone.ChainId, callbackArgs.HostZoneId, "host zone in callback args (%s)", callbackKey)
	s.Require().Equal(tc.expectedReinvestment, callbackArgs.ReinvestAmount, "reinvestment coin in callback args (%s)", callbackKey)
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_ZeroBalance() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_CommissionOverflow() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_InvalidCommission() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_SafetyCheckFailed() {
	// Not sure if this is possible to test
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_FailedSubmitTx() {
	// Remove connectionId
}
