package keeper_test

import (
	"math/rand"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	minttypes "github.com/Stride-Labs/stride/v14/x/mint/types"
	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type UpdateRedemptionRateTestCase struct {
	totalDelegation       sdkmath.Int
	undelegatedBal        sdkmath.Int
	justDepositedNative   sdkmath.Int
	justDepositedLSM      sdkmath.Int
	stSupply              sdkmath.Int
	initialRedemptionRate sdk.Dec
}

// Helper function to look up the redemption rate and check it against expectations
func (s *KeeperTestSuite) checkRedemptionRateAfterUpdate(expectedRedemptionRate sdk.Dec) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found but was not")
	s.Require().Equal(expectedRedemptionRate, hostZone.RedemptionRate, "redemption rate")
}

func (s *KeeperTestSuite) SetupUpdateRedemptionRates(tc UpdateRedemptionRateTestCase) []recordtypes.DepositRecord {
	// add some deposit records with status STAKE
	//    to comprise the undelegated delegation account balance i.e. "to be staked"
	toBeStakedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: HostChainId,
		Amount:     tc.undelegatedBal,
		Status:     recordtypes.DepositRecord_DELEGATION_QUEUE,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeStakedDepositRecord)

	// add a balance to the stakeibc module account (via records)
	//    to comprise the stakeibc deposit account balance i.e. "to be transferred"
	toBeTransferedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: HostChainId,
		Amount:     tc.justDepositedNative,
		Status:     recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeTransferedDepositRecord)

	// add an LSMTokenDeposit to represent an LSMLiquidStake that has not yet been detokenized
	lsmTokenDeposit := recordtypes.LSMTokenDeposit{
		ChainId:          HostChainId,
		Amount:           tc.justDepositedLSM,
		Status:           recordtypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		ValidatorAddress: ValAddress,
	}
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, lsmTokenDeposit)

	// set the stSupply by minting
	supply := sdk.NewCoins(sdk.NewCoin(StAtom, tc.stSupply))
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, supply)
	s.Require().NoError(err)

	// set the staked balance on the host zone
	hostZone := types.HostZone{
		ChainId:          HostChainId,
		HostDenom:        Atom,
		TotalDelegations: tc.totalDelegation,
		RedemptionRate:   tc.initialRedemptionRate,
		Validators:       []*types.Validator{{Address: ValAddress, SharesToTokensRate: sdk.OneDec()}},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return []recordtypes.DepositRecord{toBeStakedDepositRecord, toBeTransferedDepositRecord}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesSuccessful() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.NewInt(2),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedNative:   sdkmath.NewInt(4),
		justDepositedLSM:      sdkmath.NewInt(5),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	// 2 + 3 + 4 + 5 / 10 = 14 / 10 = 1.4
	expectedNewRate := sdk.MustNewDecFromStr("1.4")
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_ZeroStAssets() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.NewInt(2),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedNative:   sdkmath.NewInt(4),
		justDepositedLSM:      sdkmath.NewInt(5),
		stSupply:              sdkmath.ZeroInt(),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedRedemptionRate := sdk.NewDec(1)
	s.checkRedemptionRateAfterUpdate(expectedRedemptionRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_ZeroNativeAssets() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.ZeroInt(),
		justDepositedNative:   sdkmath.ZeroInt(),
		justDepositedLSM:      sdkmath.ZeroInt(),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedRedemptionRate := sdk.ZeroDec()
	s.checkRedemptionRateAfterUpdate(expectedRedemptionRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoDepositAccountRecords() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.NewInt(3),
		undelegatedBal:        sdkmath.NewInt(4),
		justDepositedNative:   sdkmath.NewInt(5), // should be ignored from filter below
		justDepositedLSM:      sdkmath.NewInt(6),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	// filter out the TRANSFER_QUEUE record from the records when updating the redemption rate
	filteredRecords := []recordtypes.DepositRecord{}
	for _, record := range depositRecords {
		if record.Status != recordtypes.DepositRecord_TRANSFER_QUEUE {
			filteredRecords = append(filteredRecords, record)
		}
	}
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, filteredRecords)

	// 3 + 4 + 6 / 10 = 13 / 10 = 1.3
	expectedNewRate := sdk.MustNewDecFromStr("1.3")
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoStakeDepositRecords() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.NewInt(3),
		undelegatedBal:        sdkmath.NewInt(4), // should be ignored from filter below
		justDepositedNative:   sdkmath.NewInt(5),
		justDepositedLSM:      sdkmath.NewInt(6),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	// filter out the DELEGATION_QUEUE record from the records when updating the redemption rate
	filteredRecords := []recordtypes.DepositRecord{}
	for _, record := range depositRecords {
		if record.Status != recordtypes.DepositRecord_DELEGATION_QUEUE {
			filteredRecords = append(filteredRecords, record)
		}
	}
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, filteredRecords)

	// 3 + 5 + 6 / 10 = 14 / 10 = 1.4
	expectedNewRate := sdk.MustNewDecFromStr("1.4")
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoTotalDelegation() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedNative:   sdkmath.NewInt(4),
		justDepositedLSM:      sdkmath.NewInt(5),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	// 3 + 4 + 5 / 10 = 12 / 10 = 1.2
	expectedNewRate := sdk.MustNewDecFromStr("1.2")
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_RandomInitialRedemptionRate() {
	genRandUintBelowMax := func(max int) int64 {
		min := int(1)
		n := 1 + rand.Intn(max-min+1)
		return int64(n)
	}

	// redemption rate random number, biased to be [1,2)
	max := 1_000_000
	initialRedemptionRate := sdk.NewDec(genRandUintBelowMax(max)).Quo(sdk.NewDec(genRandUintBelowMax(max / 2)))

	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.NewInt(2),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedNative:   sdkmath.NewInt(4),
		justDepositedLSM:      sdkmath.NewInt(5),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: initialRedemptionRate,
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	// 2 + 3 + 4 + 5 / 10 = 14 / 10 = 1.4
	expectedNewRate := sdk.MustNewDecFromStr("1.4")
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

// Tests GetDepositAccountBalance and GetUndelegatedBalance
func (s *KeeperTestSuite) TestGetRedemptionRate_DepositRecords() {
	// Build combinations of transfer deposit records
	toBeTransferedDepositRecords := []recordtypes.DepositRecord{
		// TRANSFER_QUEUE Total: 1 + 2 + 3 = 6
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_TRANSFER_QUEUE, Amount: sdk.NewInt(1)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_TRANSFER_QUEUE, Amount: sdk.NewInt(2)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_TRANSFER_QUEUE, Amount: sdk.NewInt(3)},

		// TRANSFER_IN_PROGRESS Total: 4 + 5 + 6 = 15
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, Amount: sdk.NewInt(4)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, Amount: sdk.NewInt(5)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, Amount: sdk.NewInt(6)},

		// Different host zone ID - should be ignored
		{HostZoneId: "different", Status: recordtypes.DepositRecord_TRANSFER_QUEUE, Amount: sdk.NewInt(1)},
		{HostZoneId: "different", Status: recordtypes.DepositRecord_TRANSFER_QUEUE, Amount: sdk.NewInt(2)},
		{HostZoneId: "different", Status: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, Amount: sdk.NewInt(4)},
		{HostZoneId: "different", Status: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, Amount: sdk.NewInt(5)},
	}
	expectedJustDepositedBalance := int64(1 + 2 + 3 + 4 + 5 + 6) // 6 + 15 = 21

	// Build combinations of delegation deposit records
	toBeStakedDepositRecords := []recordtypes.DepositRecord{
		// DELEGATION_QUEUE Total: 7 + 8 + 9 = 24
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_DELEGATION_QUEUE, Amount: sdk.NewInt(7)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_DELEGATION_QUEUE, Amount: sdk.NewInt(8)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_DELEGATION_QUEUE, Amount: sdk.NewInt(9)},

		// DELEGATION_IN_PROGRESS Total: 10 + 11 + 12 = 33
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS, Amount: sdk.NewInt(10)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS, Amount: sdk.NewInt(11)},
		{HostZoneId: HostChainId, Status: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS, Amount: sdk.NewInt(12)},

		// Different host zone ID - should be ignored
		{HostZoneId: "different", Status: recordtypes.DepositRecord_DELEGATION_QUEUE, Amount: sdk.NewInt(7)},
		{HostZoneId: "different", Status: recordtypes.DepositRecord_DELEGATION_QUEUE, Amount: sdk.NewInt(8)},
		{HostZoneId: "different", Status: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS, Amount: sdk.NewInt(10)},
		{HostZoneId: "different", Status: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS, Amount: sdk.NewInt(11)},
	}
	expectedUndelegatedBalance := int64(7 + 8 + 9 + 10 + 11 + 12) // 24 + 33 = 57

	// Use concatenation of all deposit records when running tests
	allDepositRecords := append(toBeTransferedDepositRecords, toBeStakedDepositRecords...)

	// Check the transfer records
	actualJustDepositedBalance := s.App.StakeibcKeeper.GetDepositAccountBalance(HostChainId, allDepositRecords)
	s.Require().Equal(expectedJustDepositedBalance, actualJustDepositedBalance.TruncateInt64(), "deposit account balance")

	// Check the delegation records
	actualUndelegatedBalance := s.App.StakeibcKeeper.GetUndelegatedBalance(HostChainId, allDepositRecords)
	s.Require().Equal(expectedUndelegatedBalance, actualUndelegatedBalance.TruncateInt64(), "undelegated balance")
}

func (s *KeeperTestSuite) TestGetTokenizedDelegation() {
	transferQueue := recordtypes.LSMTokenDeposit_TRANSFER_QUEUE
	transferInProgress := recordtypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS
	detokenizationQueue := recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE
	detokenizationInProgress := recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS
	transferFailed := recordtypes.LSMTokenDeposit_TRANSFER_FAILED
	detokenizationFailed := recordtypes.LSMTokenDeposit_DETOKENIZATION_FAILED

	validators := []*types.Validator{
		{Address: "valA", SharesToTokensRate: sdk.OneDec()},
		{Address: "valB", SharesToTokensRate: sdk.MustNewDecFromStr("0.75")},
		{Address: "valC", SharesToTokensRate: sdk.MustNewDecFromStr("0.5")},
	}

	// Total: 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 = 65
	lsmDeposits := []recordtypes.LSMTokenDeposit{
		// ValidatorA SharesToTokens Rate 1.0
		{ChainId: HostChainId, Status: transferInProgress, Amount: sdk.NewInt(1), ValidatorAddress: "valA"}, // 1 * 1.0 = 1
		{ChainId: HostChainId, Status: transferInProgress, Amount: sdk.NewInt(2), ValidatorAddress: "valA"}, // 2 * 1.0 = 2

		{ChainId: HostChainId, Status: detokenizationInProgress, Amount: sdk.NewInt(3), ValidatorAddress: "valA"}, // 3 * 1.0 = 3
		{ChainId: HostChainId, Status: detokenizationInProgress, Amount: sdk.NewInt(4), ValidatorAddress: "valA"}, // 4 * 1.0 = 4

		// ValidatorB SharesToTokens Rate 0.75
		{ChainId: HostChainId, Status: transferQueue, Amount: sdk.NewInt(7), ValidatorAddress: "valB"}, // 7 * 0.75 = 5.25 (5)
		{ChainId: HostChainId, Status: transferQueue, Amount: sdk.NewInt(9), ValidatorAddress: "valB"}, // 9 * 0.75 = 6.75 (6)

		{ChainId: HostChainId, Status: detokenizationQueue, Amount: sdk.NewInt(10), ValidatorAddress: "valB"}, // 10 * 0.75 = 7.5 (7)
		{ChainId: HostChainId, Status: detokenizationQueue, Amount: sdk.NewInt(11), ValidatorAddress: "valB"}, // 11 * 0.75 = 8.25 (8)

		// ValidatorC SharesToTokens Rate 0.50
		{ChainId: HostChainId, Status: transferFailed, Amount: sdk.NewInt(18), ValidatorAddress: "valC"}, // 18 * 0.5 = 9
		{ChainId: HostChainId, Status: transferFailed, Amount: sdk.NewInt(20), ValidatorAddress: "valC"}, // 20 * 0.5 = 10

		{ChainId: HostChainId, Status: detokenizationFailed, Amount: sdk.NewInt(22), ValidatorAddress: "valC"}, // 22 * 0.5 = 11
		{ChainId: HostChainId, Status: detokenizationFailed, Amount: sdk.NewInt(24), ValidatorAddress: "valC"}, // 24 * 0.5 = 12

		// Status DEPOSIT_PENDING - should be ignored
		{ChainId: HostChainId, Status: recordtypes.LSMTokenDeposit_DEPOSIT_PENDING, Amount: sdk.NewInt(11)},
		{ChainId: HostChainId, Status: recordtypes.LSMTokenDeposit_DEPOSIT_PENDING, Amount: sdk.NewInt(12)},

		// Different chain ID - should be ignored
		{ChainId: "different", Status: transferInProgress, Amount: sdk.NewInt(1)},
		{ChainId: "different", Status: detokenizationQueue, Amount: sdk.NewInt(3)},
		{ChainId: "different", Status: detokenizationInProgress, Amount: sdk.NewInt(5)},
		{ChainId: "different", Status: transferFailed, Amount: sdk.NewInt(7)},
		{ChainId: "different", Status: detokenizationFailed, Amount: sdk.NewInt(9)},

		// Non-existent validator  - should be ignored
		{ChainId: HostChainId, Status: transferInProgress, Amount: sdk.NewInt(1)},
		{ChainId: HostChainId, Status: detokenizationQueue, Amount: sdk.NewInt(3)},
		{ChainId: HostChainId, Status: detokenizationInProgress, Amount: sdk.NewInt(5)},
		{ChainId: HostChainId, Status: transferFailed, Amount: sdk.NewInt(7)},
		{ChainId: HostChainId, Status: detokenizationFailed, Amount: sdk.NewInt(9)},
	}
	expectedTokenizedDelegation := int64(1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 + 11 + 12)

	// Store deposits
	for i, deposit := range lsmDeposits {
		deposit.Denom = strconv.Itoa(i)
		s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)
	}

	// Check the total delegation from LSM Tokens
	hostZone := types.HostZone{ChainId: HostChainId, Validators: validators}
	actualTokenizedDelegation := s.App.StakeibcKeeper.GetTotalTokenizedDelegations(s.Ctx, hostZone)
	s.Require().Equal(expectedTokenizedDelegation, actualTokenizedDelegation.TruncateInt64())
}
