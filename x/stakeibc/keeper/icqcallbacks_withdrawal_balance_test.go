package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type WithdrawalBalanceCallbackState struct {
	hostZone stakeibctypes.HostZone
}

type WithdrawalBalanceCallbackArgs struct {
	args  []byte
	query icqtypes.Query
}

type WithdrawalBalanceCallbackTestCase struct {
	initialState WithdrawalBalanceCallbackState
	validArgs    WithdrawalBalanceCallbackArgs
}

func (s *KeeperTestSuite) CreateBalanceQueryArgs(address string, denom string) []byte {
	_, addressBz, err := bech32.DecodeAndConvert(address)
	s.Require().NoError(err)

	denomBz := []byte(denom)
	balancePrefix := banktypes.CreateAccountBalancesPrefix(addressBz)
	return append(balancePrefix, denomBz...)
}

func (s *KeeperTestSuite) SetupWithdrawalBalanceCallbackTest() WithdrawalBalanceCallbackTestCase {
	// Register ICA accounts (REVISIT IF THIS IS NECESSARY)
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]

	feeAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "FEE")
	s.CreateICAChannel(feeAccountOwner)
	feeAddress := s.IcaAddresses[feeAccountOwner]

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		DelegationAccount: &stakeibctypes.ICAAccount{Address: delegationAddress},
		WithdrawalAccount: &stakeibctypes.ICAAccount{Address: withdrawalAddress},
		FeeAccount:        &stakeibctypes.ICAAccount{Address: feeAddress},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	queryArgs := s.CreateBalanceQueryArgs(withdrawalAddress, Atom)

	return WithdrawalBalanceCallbackTestCase{
		initialState: WithdrawalBalanceCallbackState{
			hostZone: hostZone,
		},
		validArgs: WithdrawalBalanceCallbackArgs{
			args:  queryArgs,
			query: icqtypes.Query{},
		},
	}
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_Successful() {
	tc := s.SetupWithdrawalBalanceCallbackTest()

	err := stakeibckeeper.WithdrawalBalanceCallback(s.App.StakeibcKeeper, s.Ctx(), tc.validArgs.args, tc.validArgs.query)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_ZeroBalance() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_InvalidCommission() {

}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_SafetyCheckFailed() {
	// Not sure if this is possible to test
}

func (s *KeeperTestSuite) TestWithdrawalBalanceCallback_FailedSubmitTx() {

}
