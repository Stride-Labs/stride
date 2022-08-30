package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RestoreInterchainAccountTestCase struct {
	validMsg stakeibc.MsgRestoreInterchainAccount
}

func (s *KeeperTestSuite) SetupRestoreInterchainAccount() RestoreInterchainAccountTestCase {
	s.CreateTransferChannel(HostChainId)

	hostZone := stakeibc.HostZone{
		ChainId:      HostChainId,
		HostDenom:    Atom,
		Bech32Prefix: GaiaPrefix,
		ConnectionId: ibctesting.FirstConnectionID,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	defaultMsg := stakeibc.MsgRestoreInterchainAccount{
		Creator:     "creatoraddress",
		ChainId:     HostChainId,
		AccountType: stakeibc.ICAAccountType_DELEGATION,
	}

	return RestoreInterchainAccountTestCase{
		validMsg: defaultMsg,
	}
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_Success() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "registered ica account successfully")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsForIncorrectHostZone() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.ChainId = "incorrectchainid"
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().Error(err, "registered ica account fails for incorrect host zone")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsIfAccountExists() {
	tc := s.SetupRestoreInterchainAccount()
	s.CreateICAChannel("GAIA.DELEGATION")
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().Error(err, "registered ica account fails when account already exists")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_SucceedsIfOtherAccountExists() {
	tc := s.SetupRestoreInterchainAccount()
	s.CreateICAChannel("GAIA.WITHDRAWAL")
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "registered ica account fails when account already exists")
}
