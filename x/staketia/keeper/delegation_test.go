package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v26/x/staketia/types"
)

var InitialDelegation = sdkmath.NewInt(1_000_000)

// ----------------------------------------------------
//	               PrepareDelegation
// ----------------------------------------------------

func (s *KeeperTestSuite) TestPrepareDelegation() {
	s.CreateTransferChannel(HostChainId)

	// Only the deposit address must be valid
	depositAddress := s.TestAccs[0]
	delegationAddress := "celestiaXXX"

	// We must use a valid IBC denom for this test
	nativeIbcDenom := s.CreateAndStoreIBCDenom(HostNativeDenom)

	// Create the host zone with relevant addresses and an IBC denom
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		DepositAddress:      depositAddress.String(),
		DelegationAddress:   delegationAddress,
		NativeTokenIbcDenom: nativeIbcDenom,
		TransferChannelId:   ibctesting.FirstChannelID,
	})

	// Fund the deposit account with tokens that will be transferred
	depositAccountBalance := sdkmath.NewInt(1_000_000)
	nativeTokensInDeposit := sdk.NewCoin(nativeIbcDenom, depositAccountBalance)
	s.FundAccount(depositAddress, nativeTokensInDeposit)

	// Get next sequence number to confirm IBC transfer
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)

	// submit prepare delegation
	epochNumber := uint64(1)
	epochDuration := time.Hour * 24
	err := s.App.StaketiaKeeper.PrepareDelegation(s.Ctx, epochNumber, epochDuration)
	s.Require().NoError(err, "no error expected when preparing delegation")

	// check that a delegation record was created
	delegationRecords := s.App.StaketiaKeeper.GetAllActiveDelegationRecords(s.Ctx)
	s.Require().Equal(1, len(delegationRecords), "number of delegation records")

	// check that the delegation record has the correct id, status, and amount
	delegationRecord := delegationRecords[0]
	s.Require().Equal(epochNumber, delegationRecord.Id, "delegation record epoch number")
	s.Require().Equal(types.TRANSFER_IN_PROGRESS, delegationRecord.Status, "delegation record status")
	s.Require().Equal(depositAccountBalance, delegationRecord.NativeAmount, "delegation record amount")

	// check that the transfer in progress record was created
	transferInProgressRecordId, found := s.App.StaketiaKeeper.GetTransferInProgressRecordId(s.Ctx, ibctesting.FirstChannelID, startSequence)
	s.Require().True(found, "transfer in progress record should have been found")
	s.Require().Equal(epochNumber, transferInProgressRecordId, "transfer in progress record ID")

	// check that the tokens were burned and the sequence number was incremented
	// (indicating that the transfer was submitted)
	endSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)
	s.Require().Equal(startSequence+1, endSequence, "sequence number should have incremented")

	nativeTokenSupply := s.App.BankKeeper.GetSupply(s.Ctx, nativeIbcDenom)
	s.Require().Zero(nativeTokenSupply.Amount.Int64(), "ibc tokens should have been burned")

	// Check that the deposit account is empty
	depositAccountBalance = s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, nativeIbcDenom).Amount
	s.Require().Zero(depositAccountBalance.Int64(), "deposit account balance should be empty")

	// Check that if we ran this again immediately, it would error because there is a transfer record in progress already
	err = s.App.StaketiaKeeper.PrepareDelegation(s.Ctx, epochNumber+1, epochDuration)
	s.Require().ErrorContains(err, "cannot prepare delegation while a transfer is in progress")

	// Remove the record and try to run it again
	s.App.StaketiaKeeper.ArchiveDelegationRecord(s.Ctx, delegationRecord)
	err = s.App.StaketiaKeeper.PrepareDelegation(s.Ctx, epochNumber+1, epochDuration)
	s.Require().NoError(err, "no error expected when preparing delegation again")

	// It should not create a new record since there is nothing to delegate
	delegationRecords = s.App.StaketiaKeeper.GetAllActiveDelegationRecords(s.Ctx)
	s.Require().Equal(0, len(delegationRecords), "there should be no delegation records")

	// Halt zone
	s.App.StaketiaKeeper.HaltZone(s.Ctx)
	err = s.App.StaketiaKeeper.PrepareDelegation(s.Ctx, epochNumber, epochDuration)
	s.Require().ErrorContains(err, "host zone is halted")
}

// ----------------------------------------------------
//	               ConfirmDelegation
// ----------------------------------------------------

func (s *KeeperTestSuite) GetDefaultDelegationRecords() []types.DelegationRecord {
	delegationRecords := []types.DelegationRecord{
		{
			Id:           1,
			NativeAmount: sdk.NewInt(1000),
			Status:       types.TRANSFER_IN_PROGRESS,
			TxHash:       "",
		},
		{
			Id:           6, // out of order to make sure this won't break anything
			NativeAmount: sdk.NewInt(6000),
			Status:       types.DELEGATION_QUEUE, // to be set
			TxHash:       "",
		},
		{
			Id:           5, // out of order to make sure this won't break anything
			NativeAmount: sdk.NewInt(5000),
			Status:       types.TRANSFER_IN_PROGRESS,
			TxHash:       "",
		},
		{
			Id:           3,
			NativeAmount: sdk.NewInt(3000),
			Status:       types.TRANSFER_FAILED,
			TxHash:       "",
		},
		{
			Id:           2,
			NativeAmount: sdk.NewInt(2000),
			Status:       types.DELEGATION_QUEUE, // to be set
			TxHash:       "",
		},
		{
			Id:           7,
			NativeAmount: sdk.NewInt(7000),
			Status:       types.TRANSFER_FAILED,
			TxHash:       ValidTxHashDefault,
		},
	}

	return delegationRecords
}

// Helper function to setup delegation records, returns a list of records
func (s *KeeperTestSuite) SetupDelegationRecords() {
	// Set Delegation Records
	delegationRecords := s.GetDefaultDelegationRecords()
	// loop through and set each record
	for _, delegationRecord := range delegationRecords {
		s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	}

	// Set staketia hostZone
	hostZone := s.initializeHostZone()
	hostZone.RemainingDelegatedBalance = InitialDelegation
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// Set stakeibc host zone with the same total delegation
	stakeibcHostZone := stakeibctypes.HostZone{
		ChainId:          types.CelestiaChainId,
		TotalDelegations: InitialDelegation,
		Halted:           false,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibcHostZone)
}

func (s *KeeperTestSuite) VerifyDelegationRecords(verifyIdentical bool, archiveIds ...uint64) {
	defaultDelegationRecords := s.GetDefaultDelegationRecords()

	hostZone := s.MustGetHostZone()

	for _, defaultDelegationRecord := range defaultDelegationRecords {
		// check if record should be archived
		shouldBeArchived := false
		for _, archiveId := range archiveIds {
			if defaultDelegationRecord.Id == archiveId {
				shouldBeArchived = true
				break
			}
		}

		// grab relevant record in store
		loadedDelegationRecord := types.DelegationRecord{}
		found := false
		if shouldBeArchived {
			loadedDelegationRecord, found = s.App.StaketiaKeeper.GetArchivedDelegationRecord(s.Ctx, defaultDelegationRecord.Id)
		} else {
			loadedDelegationRecord, found = s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, defaultDelegationRecord.Id)
		}
		s.Require().True(found)
		// verify record is correct
		s.Require().Equal(defaultDelegationRecord.Id, loadedDelegationRecord.Id)
		s.Require().Equal(defaultDelegationRecord.NativeAmount, loadedDelegationRecord.NativeAmount)

		// Verify status and txHash are correct, if needed
		if (defaultDelegationRecord.Status == types.TRANSFER_FAILED) ||
			(defaultDelegationRecord.Status == types.TRANSFER_IN_PROGRESS) ||
			verifyIdentical {
			s.Require().Equal(defaultDelegationRecord.Status, loadedDelegationRecord.Status)
			s.Require().Equal(defaultDelegationRecord.TxHash, loadedDelegationRecord.TxHash)
		}

		// if nothing should have changed, verify that host zone balance is unmodified
		if verifyIdentical {
			// verify hostZone delegated balance is same as initial delegation
			s.Require().Equal(InitialDelegation.Int64(), hostZone.RemainingDelegatedBalance.Int64(), "hostZone delegated balance should not have changed")
		}
	}
}

func (s *KeeperTestSuite) TestConfirmDelegation_Successful() {
	s.SetupDelegationRecords()

	// we're halting the zone to test that the tx works even when the host zone is halted
	s.App.StaketiaKeeper.HaltZone(s.Ctx)

	// try setting valid delegation queue
	err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyDelegationRecords(false, 6)

	// verify record 6 modified
	loadedDelegationRecord, found := s.App.StaketiaKeeper.GetArchivedDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	s.Require().Equal(types.DELEGATION_COMPLETE, loadedDelegationRecord.Status, "delegation record should be updated to status DELEGATION_ARCHIVE")
	s.Require().Equal(ValidTxHashNew, loadedDelegationRecord.TxHash, "delegation record should be updated with txHash")

	// verify hostZone delegated balance is same as initial delegation + 6000
	expectedDelegation := InitialDelegation.Int64() + 6000

	hostZone := s.MustGetHostZone()
	s.Require().Equal(expectedDelegation, hostZone.RemainingDelegatedBalance.Int64(), "staketia remaining delegated balance")

	stakeibcHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, types.CelestiaChainId)
	s.Require().True(found)
	s.Require().Equal(expectedDelegation, stakeibcHostZone.TotalDelegations.Int64(), "stakeibc total delegations")
}

func (s *KeeperTestSuite) TestConfirmDelegation_DelegationZero() {
	s.SetupDelegationRecords()

	// try setting delegation queue with zero delegation
	delegationRecord, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	delegationRecord.NativeAmount = sdk.NewInt(0)
	s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState, "not allowed to confirm zero delegation")
}

func (s *KeeperTestSuite) TestConfirmDelegation_DelegationNegative() {
	s.SetupDelegationRecords()

	// try setting delegation queue with negative delegation
	delegationRecord, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	delegationRecord.NativeAmount = sdk.NewInt(-10)
	s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState, "not allowed to confirm negative delegation")
}

func (s *KeeperTestSuite) TestConfirmDelegation_RecordDoesntExist() {
	s.SetupDelegationRecords()

	// try setting invalid record id
	err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, 15, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrDelegationRecordNotFound)

	// verify delegation records haven't changed
	s.VerifyDelegationRecords(true)
}

func (s *KeeperTestSuite) TestConfirmDelegation_RecordIncorrectState() {
	s.SetupDelegationRecords()

	// first verify records in wrong status
	ids := []uint64{1, 3, 5, 7}
	for _, id := range ids {
		err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, id, ValidTxHashNew, ValidOperator)
		s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState)
		// verify delegation records haven't changed
		s.VerifyDelegationRecords(true)
	}
}
