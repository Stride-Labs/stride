package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

const (
	InitialDelegation = int64(1_000_000)
)

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
			Status:       types.DELEGATION_ARCHIVE,
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
			Status:       types.DELEGATION_ARCHIVE,
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

	// Set HostZone
	hostZone := s.initializeHostZone()
	hostZone.DelegatedBalance = sdk.NewInt(InitialDelegation)
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)
}

func (s *KeeperTestSuite) VerifyDelegationRecords(verifyIdentical bool, archiveIds ...uint64) {
	defaultDelegationRecords := s.GetDefaultDelegationRecords()

	hostZone := s.MustGetHostZone()

	for _, defaultDelegationRecord := range defaultDelegationRecords {
		// check if record should be archived
		shouldBeArchived := false
		if defaultDelegationRecord.Status == types.DELEGATION_ARCHIVE {
			shouldBeArchived = true
		} else {
			for _, archiveId := range archiveIds {
				if defaultDelegationRecord.Id == archiveId {
					shouldBeArchived = true
					break
				}
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
		if (defaultDelegationRecord.Status == types.DELEGATION_ARCHIVE) ||
			(defaultDelegationRecord.Status == types.TRANSFER_IN_PROGRESS) ||
			verifyIdentical {
			s.Require().Equal(defaultDelegationRecord.Status, loadedDelegationRecord.Status)
			s.Require().Equal(defaultDelegationRecord.TxHash, loadedDelegationRecord.TxHash)
		}

		// if nothing should have changed, verify that host zone balance is unmodified
		if verifyIdentical {
			// verify hostZone delegated balance is same as initial delegation
			s.Require().Equal(InitialDelegation, hostZone.DelegatedBalance.Int64(), "hostZone delegated balance should not have changed")
		}
	}
}

func (s *KeeperTestSuite) TestConfirmDelegation_Successful() {
	s.SetupDelegationRecords()

	// try setting valid delegation queue
	err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyDelegationRecords(false, 6)

	// verify record 6 modified
	loadedDelegationRecord, found := s.App.StaketiaKeeper.GetArchivedDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	s.Require().Equal(types.DELEGATION_ARCHIVE, loadedDelegationRecord.Status, "delegation record should be updated to status DELEGATION_ARCHIVE")
	s.Require().Equal(ValidTxHashNew, loadedDelegationRecord.TxHash, "delegation record should be updated with txHash")

	// verify hostZone delegated balance is same as initial delegation + 6000
	hostZone := s.MustGetHostZone()
	s.Require().Equal(InitialDelegation+6000, hostZone.DelegatedBalance.Int64(), "hostZone delegated balance should have increased by 6000")
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
	ids := []uint64{1, 5}
	for _, id := range ids {
		err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, id, ValidTxHashNew, ValidOperator)
		s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState)
		// verify delegation records haven't changed
		s.VerifyDelegationRecords(true)
	}

	// then verify archived records
	ids = []uint64{3, 7}
	for _, id := range ids {
		err := s.App.StaketiaKeeper.ConfirmDelegation(s.Ctx, id, ValidTxHashNew, ValidOperator)
		s.Require().ErrorIs(err, types.ErrDelegationRecordNotFound)
		// verify delegation records haven't changed
		s.VerifyDelegationRecords(true)
	}
}
