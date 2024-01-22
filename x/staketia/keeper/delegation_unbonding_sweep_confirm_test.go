package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// define constants
const (
	DefaultClaimFundingAmount = 2600 // sum of NativeTokenAmount of records with status UNBONDED
)

func (s *KeeperTestSuite) GetDefaultUnbondingRecords() []types.UnbondingRecord {
	unbondingRecords := []types.UnbondingRecord{ // verify no issue if these are out of order
		{
			Id:                             1,
			Status:                         types.UNBONDING_QUEUE,
			StTokenAmount:                  sdk.NewInt(100),
			NativeAmount:                   sdk.NewInt(200),
			UnbondingCompletionTimeSeconds: 0,
			UndelegationTxHash:             "",
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             7,
			Status:                         types.CLAIMABLE,
			StTokenAmount:                  sdk.NewInt(200),
			NativeAmount:                   sdk.NewInt(400),
			UnbondingCompletionTimeSeconds: 10,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       ValidTxHashDefault,
		},
		{
			Id:                             5,
			Status:                         types.UNBONDING_IN_PROGRESS,
			StTokenAmount:                  sdk.NewInt(500),
			NativeAmount:                   sdk.NewInt(1000),
			UnbondingCompletionTimeSeconds: 20,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             3,
			Status:                         types.ACCUMULATING_REDEMPTIONS,
			StTokenAmount:                  sdk.NewInt(300),
			NativeAmount:                   sdk.NewInt(600),
			UnbondingCompletionTimeSeconds: 0,
			UndelegationTxHash:             "",
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             6,
			Status:                         types.UNBONDED,
			StTokenAmount:                  sdk.NewInt(600),
			NativeAmount:                   sdk.NewInt(1200),
			UnbondingCompletionTimeSeconds: 15,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             4,
			Status:                         types.UNBONDING_ARCHIVE,
			StTokenAmount:                  sdk.NewInt(400),
			NativeAmount:                   sdk.NewInt(800),
			UnbondingCompletionTimeSeconds: 5,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       ValidTxHashDefault,
		},
		{
			Id:                             2,
			Status:                         types.UNBONDED,
			StTokenAmount:                  sdk.NewInt(700),
			NativeAmount:                   sdk.NewInt(1400),
			UnbondingCompletionTimeSeconds: 18,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       "",
		},
	}
	return unbondingRecords
}

// Helper function to setup unbonding records, returns a list of records
func (s *KeeperTestSuite) SetupUnbondingRecords(amount int64) {
	unbondingRecords := s.GetDefaultUnbondingRecords()

	// loop through and set each record
	for _, unbondingRecord := range unbondingRecords {
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}

	// setup host zone, to fund claim address
	hostZone := s.initializeHostZone()
	hostZone.ClaimAddress = "stride1njt6kn0c2a2w5ax8mlm9k0fmcc8tyjgh7s8hu8" // random address
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// fund claim address
	claimAddress, err := sdk.AccAddressFromBech32(hostZone.ClaimAddress)
	s.Require().NoError(err)
	liquidStakeToken := sdk.NewCoin(hostZone.NativeTokenIbcDenom, sdk.NewInt(amount))
	s.FundAccount(claimAddress, liquidStakeToken)
}

func (s *KeeperTestSuite) VerifyUnbondingRecords(verifyUpdatedFieldsIdentical bool) {
	defaultUnbondingRecords := s.GetDefaultUnbondingRecords()
	for _, defaultUnbondingRecord := range defaultUnbondingRecords {
		// grab relevant record in store
		loadedUnbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, defaultUnbondingRecord.Id)
		s.Require().True(found)

		// verify record is correct
		s.Require().Equal(defaultUnbondingRecord.Id, loadedUnbondingRecord.Id)
		s.Require().Equal(defaultUnbondingRecord.NativeAmount, loadedUnbondingRecord.NativeAmount)
		s.Require().Equal(defaultUnbondingRecord.StTokenAmount, loadedUnbondingRecord.StTokenAmount)
		s.Require().Equal(defaultUnbondingRecord.UnbondingCompletionTimeSeconds, loadedUnbondingRecord.UnbondingCompletionTimeSeconds)
		s.Require().Equal(defaultUnbondingRecord.UndelegationTxHash, loadedUnbondingRecord.UndelegationTxHash)

		// if relevant, verify status and tx hash
		if (defaultUnbondingRecord.Status != types.UNBONDED) ||
			verifyUpdatedFieldsIdentical {
			s.Require().Equal(defaultUnbondingRecord.Status, loadedUnbondingRecord.Status)
			s.Require().Equal(defaultUnbondingRecord.UnbondedTokenSweepTxHash, loadedUnbondingRecord.UnbondedTokenSweepTxHash)
		}
	}
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_Successful() {
	s.SetupUnbondingRecords(DefaultClaimFundingAmount)

	// process record 6
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyUnbondingRecords(false)

	// verify record 6 modified
	loadedUnbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, 6)
	s.Require().True(found)
	s.Require().Equal(types.CLAIMABLE, loadedUnbondingRecord.Status, "unbonding record should be updated to status CLAIMABLE")
	s.Require().Equal(ValidTxHashNew, loadedUnbondingRecord.UnbondedTokenSweepTxHash, "unbonding record should be updated with token sweep txHash")

	// process record 2
	err = s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 2, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyUnbondingRecords(false)

	// verify record 2 modified
	loadedUnbondingRecord, found = s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, 2)
	s.Require().True(found)
	s.Require().Equal(types.CLAIMABLE, loadedUnbondingRecord.Status, "unbonding record should be updated to status CLAIMABLE")
	s.Require().Equal(ValidTxHashNew, loadedUnbondingRecord.UnbondedTokenSweepTxHash, "unbonding record should be updated with token sweep txHash")
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_NotFunded() {
	s.SetupUnbondingRecords(10)

	// try setting with no hash
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrInsufficientFunds, "should error when claim account doesn't have enough funds")

	// verify delegation records haven't changed
	s.VerifyUnbondingRecords(true)
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_InvalidClaimAddress() {
	s.SetupUnbondingRecords(DefaultClaimFundingAmount)

	hostZone := s.MustGetHostZone()
	hostZone.ClaimAddress = "strideinvalidaddress" // random address
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// try setting with no hash
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorContains(err, "decoding bech32 failed", "should error when claim address is invalid")

	// verify delegation records haven't changed
	s.VerifyUnbondingRecords(true)
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_RecordDoesntExist() {
	s.SetupUnbondingRecords(DefaultClaimFundingAmount)

	// process record 15
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 15, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrUnbondingRecordNotFound, "should error when record doesn't exist")
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_RecordIncorrectState() {
	s.SetupUnbondingRecords(DefaultClaimFundingAmount)

	// get list of ids to try
	ids := []uint64{1, 3, 4, 5, 7}
	for _, id := range ids {
		err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, id, ValidTxHashNew, ValidOperator)
		s.Require().ErrorIs(err, types.ErrInvalidUnbondingRecord, "should error when record is in incorrect state")

		// verify unbonding records haven't changed
		s.VerifyUnbondingRecords(true)
	}
}
