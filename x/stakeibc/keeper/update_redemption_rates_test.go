package keeper_test

import (
	// "fmt"

	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"

	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type UpdateRedemptionRatesTestCase struct {
	hostZone   stakeibctypes.HostZone
	allRecords []recordtypes.DepositRecord
}

func (s *KeeperTestSuite) SetupUpdateRedemptionRates(
	stakedBal sdk.Int,
	undelegatedBal sdk.Int,
	justDepositedBal sdk.Int,
	stSupply sdk.Int,
	initialRedemptionRate sdk.Dec,
) UpdateRedemptionRatesTestCase {
	// add some deposit records with status STAKE
	//    to comprise the undelegated delegation account balance i.e. "to be staked"
	toBeStakedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: "GAIA",
		Amount:     undelegatedBal,
		Status:     recordtypes.DepositRecord_DELEGATION_QUEUE,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeStakedDepositRecord)

	// add a balance to the stakeibc module account (via records)
	//    to comprise the stakeibc module account balance i.e. "to be transferred"
	toBeTransferedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: "GAIA",
		Amount:     justDepositedBal,
		Status:     recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeTransferedDepositRecord)

	// set the stSupply by minting to a random user account
	user := Account{
		acc:           s.TestAccs[0],
		stAtomBalance: sdk.NewCoin(StAtom, stSupply),
	}
	s.FundAccount(user.acc, user.stAtomBalance)

	// set the staked balance on the host zone
	hostZone := stakeibctypes.HostZone{
		ChainId:        "GAIA",
		HostDenom:      "uatom",
		StakedBal:      stakedBal,
		RedemptionRate: initialRedemptionRate,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return UpdateRedemptionRatesTestCase{
		hostZone:   hostZone,
		allRecords: []recordtypes.DepositRecord{toBeStakedDepositRecord, toBeTransferedDepositRecord},
	}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesSuccessful() {
	stakedBal := sdk.NewInt(5)
	undelegatedBal := sdk.NewInt(3)
	justDepositedBal := sdk.NewInt(3)
	stSupply := sdk.NewInt(10)

	initialRedemptionRate := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1), "t0 rr")

	records := tc.allRecords
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	expectedNewRate := sdk.NewDec(5 + 3 + 3).Quo(sdk.NewDec(10))
	s.Require().Equal(rrNew, expectedNewRate, "rr as expected")
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesRandomized() {
	// run N tests, each with random inputs

	MAX := "1_000_000_000"
	stakedBal, _ := sdk.NewIntFromString(MAX)
	undelegatedBal, _ := sdk.NewIntFromString(MAX)
	justDepositedBal, _ := sdk.NewIntFromString(MAX)
	stSupply, _ := sdk.NewIntFromString(MAX)

	// s.Require().ElementsMatch([]int{0, 0, 0, 0}, []int{int(stakedBal), int(undelegatedBal), int(justDepositedBal), int(stSupply)}) //
	initialRedemptionRate := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1), "t0 rr")

	records := tc.allRecords
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	numerator := stakedBal.Add(undelegatedBal).Add(justDepositedBal)
	denominator := stSupply
	expectedNewRate := sdk.NewDecFromInt(numerator.Quo(denominator))

	componentDescription := fmt.Sprintf(
		"Components - StakedBal: %v, UndelegateBalance: %v, JustDepositedBalance: %v, stSupply: %v, InitialRedemptionRate: %v",
		stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	s.Require().Equal(rrNew, expectedNewRate,
		"ExpectedRedemptionRate: %v, ActualRedemptionRate: %v; %s", expectedNewRate, rrNew, componentDescription)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesRandomized_MultipleRuns() {
	for i := 0; i < 100; i++ {
		s.TestUpdateRedemptionRatesRandomized()
		// reset the testing app between runs
		s.Setup()
	}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateZeroStAssets() {
	stakedBal := sdk.NewInt(5)
	undelegatedBal := sdk.NewInt(3)
	justDepositedBal := sdk.NewInt(3)
	stSupply := sdk.NewInt(0)

	initialRedemptionRate := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1), "t0 rr")

	records := tc.allRecords
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	// RR should be unchanged
	s.Require().Equal(rrNew, sdk.NewDec(1), "rr as expected")
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateZeroNativeAssets() {
	stakedBal := sdk.NewInt(0)
	undelegatedBal := sdk.NewInt(0)
	justDepositedBal := sdk.NewInt(0)
	stSupply := sdk.NewInt(10)

	initialRedemptionRate := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1), "t0 rr")

	records := tc.allRecords
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	// RR should be 0
	s.Require().Equal(rrNew, sdk.NewDec(0), "rr as expected")
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateNoModuleAccountRecords() {
	stakedBal := sdk.NewInt(5)
	undelegatedBal := sdk.NewInt(3)
	justDepositedBal := sdk.NewInt(3)
	stSupply := sdk.NewInt(10)
	initialRedemptionRate := sdk.NewDec(1)

	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1), "t0 rr")

	// filter out the TRANSFER_QUEUE record from the records when updating the redemption rate
	records := tc.allRecords[1:]
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	expectedNewRate := sdk.NewDec(5 + 3).Quo(sdk.NewDec(10))
	s.Require().Equal(rrNew, expectedNewRate, "rr as expected")
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateNoStakeDepositRecords() {
	stakedBal := sdk.NewInt(5)
	undelegatedBal := sdk.NewInt(3)
	justDepositedBal := sdk.NewInt(3)
	stSupply := sdk.NewInt(10)
	initialRedemptionRate := sdk.NewDec(1)

	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1), "t0 rr")

	// filter out the DELEGATION_QUEUE record from the records when updating the redemption rate
	records := tc.allRecords[:1]
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	numerator := stakedBal.Add(justDepositedBal)
	denominator := stSupply
	expectedNewRate := sdk.NewDecFromInt(numerator).Quo(sdk.NewDecFromInt(denominator))
	s.Require().Equal(rrNew, expectedNewRate, "rr as expected")
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateNoStakedBal() {
	undelegatedBal := sdk.NewInt(3)
	justDepositedBal := sdk.NewInt(3)
	stSupply := sdk.NewInt(10)
	initialRedemptionRate := sdk.NewDec(1)

	// SET HZ STAKED BAL TO 0
	tc := s.SetupUpdateRedemptionRates(sdk.ZeroInt(), undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(initialRedemptionRate, sdk.NewDec(1))

	records := tc.allRecords
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	expectedNewRate := sdk.NewDec(3 + 3).Quo(sdk.NewDec(10))
	s.Require().Equal(rrNew, expectedNewRate, "rr as expected")
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateRandominitialRedemptionRate() {
	stakedBal := sdk.NewInt(5)
	undelegatedBal := sdk.NewInt(3)
	justDepositedBal := sdk.NewInt(3)
	stSupply := sdk.NewInt(10)

	genRandUintBelowMax := func(MAX int) int64 {
		MIN := int(1)
		n := 1 + rand.Intn(MAX-MIN+1)
		return int64(n)
	}

	MAX := 1_000_000
	// redemption rate random number, biased to be [1,2)
	initialRedemptionRate := sdk.NewDec(genRandUintBelowMax(MAX)).Quo(sdk.NewDec(genRandUintBelowMax(MAX / 2)))

	// SET HZ STAKED BAL TO 0
	tc := s.SetupUpdateRedemptionRates(stakedBal, undelegatedBal, justDepositedBal, stSupply, initialRedemptionRate)

	records := tc.allRecords
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "hz found")
	rrNew := hz.RedemptionRate

	expectedNewRate := sdk.NewDec(3 + 3 + 5).Quo(sdk.NewDec(10))
	s.Require().Equal(rrNew, expectedNewRate, "rr as expected")
}
