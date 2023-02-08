package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v5/x/records/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	_ "github.com/stretchr/testify/suite"
)

type SubmitTxTestcase struct {
	hostZone      stakeibctypes.HostZone
	amt           sdk.Coin
	depositRecord []recordstypes.DepositRecord
	msgs          []sdk.Msg
	epochTracker  stakeibctypes.EpochTracker
}

func (s *KeeperTestSuite) SetupSubmitTx_emptyStrideEpoch() SubmitTxTestcase {
	stakedBal := sdk.NewInt(5_000)
	//Set Deposit Records of type Delegate
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]

	DepositRecordDelegate := []recordstypes.DepositRecord{
		{
			Id:         1,
			Amount:     sdk.NewInt(1000),
			Denom:      Atom,
			HostZoneId: HostChainId,
			Status:     recordstypes.DepositRecord_DELEGATION_QUEUE,
		},
		{
			Id:         2,
			Amount:     sdk.NewInt(3000),
			Denom:      Atom,
			HostZoneId: HostChainId,
			Status:     recordstypes.DepositRecord_DELEGATION_QUEUE,
		},
	}
	stakeAmount := sdk.NewInt(1_000_000)
	stakeCoin := sdk.NewCoin(Atom, stakeAmount)

	//  define the host zone with stakedBal and validators with staked amounts
	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	amtVal1 := sdk.NewInt(1_000_000)
	amtVal2 := sdk.NewInt(2_000_000)
	wgtVal1 := uint64(1)
	wgtVal2 := uint64(2)

	validators := []*stakeibctypes.Validator{
		{
			Address:       hostVal1Addr,
			DelegationAmt: amtVal1,
			Weight:        wgtVal1,
		},
		{
			Address:       hostVal2Addr,
			DelegationAmt: amtVal2,
			Weight:        wgtVal2,
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		StakedBal:      stakedBal,
		Validators:     validators,
		DelegationAccount: &stakeibctypes.ICAAccount{
			Address: delegationAddress,
			Target:  stakeibctypes.ICAAccountType_DELEGATION,
		},
		WithdrawalAccount: &stakeibctypes.ICAAccount{
			Address: withdrawalAddress,
			Target:  stakeibctypes.ICAAccountType_WITHDRAWAL,
		},
		ConnectionId: ibctesting.FirstConnectionID,
	}

	delegationIca := hostZone.DelegationAccount
	withdrawalIcaAddr := hostZone.WithdrawalAccount.Address
	msgs := []sdk.Msg{
		&distributiontypes.MsgSetWithdrawAddress{
			DelegatorAddress: delegationIca.Address,
			WithdrawAddress:  withdrawalIcaAddr,
		},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return SubmitTxTestcase{
		hostZone:      hostZone,
		amt:           stakeCoin,
		depositRecord: DepositRecordDelegate,
		msgs:          msgs,
	}
}

func (s *KeeperTestSuite) SetupSubmitTx() SubmitTxTestcase {
	tc := s.SetupSubmitTx_emptyStrideEpoch()
	//set Stride epoch
	epochIdentifier := epochtypes.STRIDE_EPOCH
	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        uint64(2),
		NextEpochStartTime: uint64(time.Now().UnixNano()),
		Duration:           uint64(2),
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	return SubmitTxTestcase{
		hostZone:      tc.hostZone,
		amt:           tc.amt,
		depositRecord: tc.depositRecord,
		msgs:          tc.msgs,
		epochTracker:  epochTracker,
	}
}

func (s *KeeperTestSuite) TestDelegateOnHost_Successful() {
	tc := s.SetupSubmitTx()
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])
	DelegateDepositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, tc.depositRecord[0].Id)
	s.Require().NoError(err)
	s.Require().Equal(DelegateDepositRecord.Status, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, "record must be status of Delegation in progress")
	s.Require().Equal(found, true, "record must be found")
}

func (s *KeeperTestSuite) TestDelegateOnHost_ConnectionIDNotFound() {
	tc := s.SetupSubmitTx()
	//change Hostzone's chainId so no connection should be found
	tc.hostZone.ChainId = "superGAIA"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])

	error := `icacontroller-superGAIA.DELEGATION has no associated connection: `
	error += `invalid chain-id`
	s.EqualError(err, error, "Hostzone's chainId has been changed so this should fail")
}

func (s *KeeperTestSuite) TestDelegateOnHost_InvalidDelegationAccount() {
	tc := s.SetupSubmitTx()
	//change Hostzone's Delegation accounts so no accounts should be found
	AddressDelegateAccount := ""
	DelegateAccount := &types.ICAAccount{Address: AddressDelegateAccount, Target: types.ICAAccountType_DELEGATION}
	tc.hostZone.DelegationAccount = DelegateAccount
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])

	error := `Invalid delegation account: `
	error += `invalid address`
	s.EqualError(err, error, "Hostzone's Delegation accounts has been changed so this should fail")
}

func (s *KeeperTestSuite) TestDelegateOnHost_ErrorGettingTargetDelegateAmtOnValidators() {
	tc := s.SetupSubmitTx()
	//change Hostzone's Validators to empty so error should be returnd
	validators := []*stakeibctypes.Validator{}
	tc.hostZone.Validators = validators
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])

	error := `no non-zero validator weights`
	s.EqualError(err, error, "Hostzone's Validators has been cleared so this should fail")
}

func (s *KeeperTestSuite) TestDelegateOnHost_FailedToGetICATimeoutNanos() {
	tc := s.SetupSubmitTx_emptyStrideEpoch()
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestUpdateWithdrawalBalance_successful() {
	tc := s.SetupSubmitTx()
	err := s.App.StakeibcKeeper.UpdateWithdrawalBalance(s.Ctx, tc.hostZone)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestUpdateWithdrawalBalance_FailedToGetICATimeoutNanos() {
	tc := s.SetupSubmitTx_emptyStrideEpoch()

	err := s.App.StakeibcKeeper.UpdateWithdrawalBalance(s.Ctx, tc.hostZone)
	expectedErr := "Failed to get ICA timeout nanos for epochType stride_epoch using param, error: "
	expectedErr += "Failed to get epoch tracker for stride_epoch: invalid request: invalid request"
	s.EqualError(err, expectedErr, "Hostzone is set without Stride Epoch so it should fail")
}

func (s *KeeperTestSuite) TestUpdateWithdrawalBalance_EmptyConnectionId() {
	tc := s.SetupSubmitTx()
	//change hostzone's connectionId
	tc.hostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	err := s.App.StakeibcKeeper.UpdateWithdrawalBalance(s.Ctx, tc.hostZone)
	expectedErr := "[ICQ Validation Check] Failed! connection id cannot be empty: "
	expectedErr += "invalid request"
	s.EqualError(err, expectedErr, "Hostzone is set without Stride Epoch so it should fail")
}

func (s *KeeperTestSuite) TestSetWithdrawalAddressOnHost_successful() {
	tc := s.SetupSubmitTx()
	err := s.App.StakeibcKeeper.SetWithdrawalAddressOnHost(s.Ctx, tc.hostZone)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestSetWithdrawalAddressOnHost_FailedToGetICATimeoutNanos() {
	tc := s.SetupSubmitTx_emptyStrideEpoch()

	err := s.App.StakeibcKeeper.SetWithdrawalAddressOnHost(s.Ctx, tc.hostZone)
	expectedErr := fmt.Sprintf("Failed to SubmitTxs for %s, %s, [%s]: ", tc.hostZone.ConnectionId, tc.hostZone.ChainId, tc.msgs[0])
	expectedErr += "invalid request"
	s.EqualError(err, expectedErr, "Hostzone is set without Stride Epoch so it should fail")
}

func (s *KeeperTestSuite) TestGetStartTimeNextEpoch_Success() {
	tc := s.SetupSubmitTx()
	epochIdentifier := epochtypes.STRIDE_EPOCH

	epochReturn, err := s.App.StakeibcKeeper.GetStartTimeNextEpoch(s.Ctx, epochIdentifier)

	s.Require().NoError(err)

	s.Require().Equal(epochReturn, tc.epochTracker.NextEpochStartTime)

}
func (s *KeeperTestSuite) TestGetStartTimeNextEpoch_FailedToGetEpoch() {
	s.SetupSubmitTx()
	//finding "epoch_stride" which is not a defined epochIdentifier
	_, err := s.App.StakeibcKeeper.GetStartTimeNextEpoch(s.Ctx, "epoch_stride")

	s.Require().EqualError(err, fmt.Sprintf("Failed to get epoch tracker for %s: %s", "epoch_stride", "invalid request"))
}

func (s *KeeperTestSuite) TestSubmitsTx_Successful() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().NoError(err)

}
func (s *KeeperTestSuite) TestSubmitsTx_FailedCallBackSendTx() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	tc.epochTracker.NextEpochStartTime = 0

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().EqualError(err, "timeout timestamp must be in the future")
}

func (s *KeeperTestSuite) TestSubmitsTx_FailedToRetrieveActiveChannel() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	account.Target = stakeibctypes.ICAAccountType_FEE

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().EqualError(err, "failed to retrieve active channel for port icacontroller-GAIA.FEE: no active channel for this owner")

}

func (s *KeeperTestSuite) TestSubmitsTx_InvalidConnectionIdNotFound() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	ConnectionID_test := "connection_test"

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, ConnectionID_test, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().EqualError(err, fmt.Sprintf("invalid connection id, %s not found", ConnectionID_test))
}
func (s *KeeperTestSuite) TestSubmitTxsEpoch_Successful() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	connectionID := ibctesting.FirstConnectionID
	epochIdentifier := epochtypes.STRIDE_EPOCH

	_, err := s.App.StakeibcKeeper.SubmitTxsEpoch(s.Ctx, connectionID, msgs, *account, epochIdentifier, "", nil)
	s.Require().NoError(err)
}
func (s *KeeperTestSuite) TestSubmitTxsEpoch_FailedCallBackSubmitTxs() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	connectionID := "connectionID_test"
	epochIdentifier := epochtypes.STRIDE_EPOCH

	_, err := s.App.StakeibcKeeper.SubmitTxsEpoch(s.Ctx, connectionID, msgs, *account, epochIdentifier, "", nil)
	s.Require().EqualError(err, fmt.Sprintf("invalid connection id, %s not found", connectionID))
}
func (s *KeeperTestSuite) TestSubmitTxsEpoch_FailedToGetICA() {
	tc := s.SetupSubmitTx()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	connectionID := ibctesting.FirstConnectionID
	//test case
	epochIdentifier := "epochtypes_test"

	_, err := s.App.StakeibcKeeper.SubmitTxsEpoch(s.Ctx, connectionID, msgs, *account, epochIdentifier, "", nil)
	s.Require().EqualError(err, fmt.Sprintf("Failed to convert timeoutNanos to uint64, error: Failed to get epoch tracker for %s: invalid request: invalid request", epochIdentifier))
}
