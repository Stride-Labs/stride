package keeper_test

import (
	// "fmt"

	"math/rand"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	minttypes "github.com/Stride-Labs/stride/v8/x/mint/types"

	recordtypes "github.com/Stride-Labs/stride/v8/x/records/types"

	stakeibctypes "github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

// TODO: Use array-of-test-cases setup for this redemption rate test
// TODO [LSM]: Fix randomized test
// TODO [LSM]: Add balanced/unbalanced delegation test cases
// TODO [LSM]: Add test cases with multiple deposit records
// TODO [LSM]: Add unit tests for sub-functions

type UpdateRedemptionRateTestCase struct {
	balancedDelegation    sdkmath.Int
	unbalancedDelegation  sdkmath.Int
	undelegatedBal        sdkmath.Int
	justDepositedBal      sdkmath.Int
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
	//    to comprise the stakeibc module account balance i.e. "to be transferred"
	toBeTransferedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: HostChainId,
		Amount:     tc.justDepositedBal,
		Status:     recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeTransferedDepositRecord)

	// set the stSupply by minting
	supply := sdk.NewCoins(sdk.NewCoin(StAtom, tc.stSupply))
	s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, supply)

	// set the staked balance on the host zone
	hostZone := stakeibctypes.HostZone{
		ChainId:                  HostChainId,
		HostDenom:                Atom,
		TotalBalancedDelegations: tc.balancedDelegation,
		RedemptionRate:           tc.initialRedemptionRate,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return []recordtypes.DepositRecord{toBeStakedDepositRecord, toBeTransferedDepositRecord}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesSuccessful() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(5),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedBal:      sdkmath.NewInt(3),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedNewRate := sdk.NewDec(5 + 3 + 3).Quo(sdk.NewDec(10))
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesRandomized() {
	// run N tests, each with random inputs
	max := int64(1_000_000_000)
	tc := UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(max),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(max),
		justDepositedBal:      sdkmath.NewInt(max),
		stSupply:              sdkmath.NewInt(max),
		initialRedemptionRate: sdk.NewDec(1),
	}
	depositRecords := s.SetupUpdateRedemptionRates(tc)

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	numerator := tc.balancedDelegation.Add(tc.undelegatedBal).Add(tc.justDepositedBal)
	denominator := tc.stSupply
	expectedNewRate := sdk.NewDecFromInt(numerator.Quo(denominator))
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesRandomized_MultipleRuns() {
	for i := 0; i < 100; i++ {
		s.TestUpdateRedemptionRatesRandomized()
		// reset the testing app between runs
		s.Setup()
	}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateZeroStAssets() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(5),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedBal:      sdkmath.NewInt(3),
		stSupply:              sdkmath.ZeroInt(),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedRedemptionRate := sdk.NewDec(1)
	s.checkRedemptionRateAfterUpdate(expectedRedemptionRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateZeroNativeAssets() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(0),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.ZeroInt(),
		justDepositedBal:      sdkmath.ZeroInt(),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedRedemptionRate := sdk.ZeroDec()
	s.checkRedemptionRateAfterUpdate(expectedRedemptionRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateNoDepositAccountRecords() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(5),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedBal:      sdkmath.NewInt(3),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	// filter out the TRANSFER_QUEUE record from the records when updating the redemption rate
	records := depositRecords[1:]
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	expectedNewRate := sdk.NewDec(5 + 3).Quo(sdk.NewDec(10))
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateNoStakeDepositRecords() {
	tc := UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(5),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedBal:      sdkmath.NewInt(3),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	}
	depositRecords := s.SetupUpdateRedemptionRates(tc)

	// filter out the DELEGATION_QUEUE record from the records when updating the redemption rate
	records := depositRecords[:1]
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	numerator := tc.balancedDelegation.Add(tc.justDepositedBal)
	denominator := tc.stSupply
	expectedNewRate := sdk.NewDecFromInt(numerator).Quo(sdk.NewDecFromInt(denominator))
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateNoBalancedDelegation() {
	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.ZeroInt(),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedBal:      sdkmath.NewInt(3),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdk.NewDec(1),
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedNewRate := sdk.NewDec(3 + 3).Quo(sdk.NewDec(10))
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateRandomInitialRedemptionRate() {
	genRandUintBelowMax := func(max int) int64 {
		min := int(1)
		n := 1 + rand.Intn(max-min+1)
		return int64(n)
	}

	// redemption rate random number, biased to be [1,2)
	max := 1_000_000
	initialRedemptionRate := sdk.NewDec(genRandUintBelowMax(max)).Quo(sdk.NewDec(genRandUintBelowMax(max / 2)))

	depositRecords := s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		balancedDelegation:    sdkmath.NewInt(5),
		unbalancedDelegation:  sdkmath.ZeroInt(),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedBal:      sdkmath.NewInt(3),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: initialRedemptionRate,
	})

	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, depositRecords)

	expectedNewRate := sdk.NewDec(3 + 3 + 5).Quo(sdk.NewDec(10))
	s.checkRedemptionRateAfterUpdate(expectedNewRate)
}
