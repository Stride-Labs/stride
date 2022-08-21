package keeper_test

import (
	// "fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	// "github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type UpdateRedemptionRatesTestCase struct {
	host_zone          stakeibc.HostZone
	redemption_rate_t0 sdk.Dec
	staked_bal         uint64
	undelegated_bal    uint64
	just_deposited_bal uint64
	st_supply          uint64
}

func (s *KeeperTestSuite) SetupUpdateRedemptionRates(
	STAKED_BAL uint64,
	UNDELEGATED_BAL uint64,
	JUST_DEPOSITED_BAL uint64,
	ST_SUPPLY uint64,
	REDEMPTION_RATE_T0 sdk.Dec,
) UpdateRedemptionRatesTestCase {

	// add some deposit records with status STAKE
	//    to comprise the undelegated delegation account balance i.e. "to be staked"
	toBeStakedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: "GAIA",
		Amount:     int64(UNDELEGATED_BAL),
		Status:     recordtypes.DepositRecord_STAKE,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeStakedDepositRecord)

	// add a balance to the stakeibc module account (via records)
	//    to comprise the stakeibc module account balance i.e. "to be transferred"
	toBeTransferedDepositRecord := recordtypes.DepositRecord{
		HostZoneId: "GAIA",
		Amount:     int64(JUST_DEPOSITED_BAL),
		Status:     recordtypes.DepositRecord_TRANSFER,
	}
	s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, toBeTransferedDepositRecord)

	// set the stSupply by minting to a random user account
	user := Account{
		acc:           s.TestAccs[0],
		stAtomBalance: sdk.NewInt64Coin(stAtom, int64(ST_SUPPLY)),
	}
	s.FundAccount(user.acc, user.stAtomBalance)

	// set the staked balance on the host zone
	hostZone := stakeibc.HostZone{
		ChainId:        "GAIA",
		HostDenom:      "uatom",
		StakedBal:      STAKED_BAL,
		RedemptionRate: REDEMPTION_RATE_T0,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return UpdateRedemptionRatesTestCase{
		host_zone:          hostZone,
		redemption_rate_t0: REDEMPTION_RATE_T0,
		staked_bal:         STAKED_BAL,
		undelegated_bal:    UNDELEGATED_BAL,
		just_deposited_bal: JUST_DEPOSITED_BAL,
		st_supply:          ST_SUPPLY,
	}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesSuccessful() {

	STAKED_BAL := uint64(5)
	UNDELEGATED_BAL := uint64(3)
	JUST_DEPOSITED_BAL := uint64(3)
	ST_SUPPLY := uint64(10)

	REDEMPTION_RATE_T0 := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(STAKED_BAL, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	s.Require().Equal(rrNew, sdk.NewDec(int64(STAKED_BAL)+int64(UNDELEGATED_BAL)+int64(JUST_DEPOSITED_BAL)).Quo(sdk.NewDec(int64(ST_SUPPLY))))
	s.Require().Equal(rrNew, sdk.NewDec(11).Quo(sdk.NewDec(10)))
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesRandomized() {
	// run N tests, each with random inputs

	// set rand seed for consistency
	genRandUintBelowMax := func(MAX int) uint64 {
		MIN := int(0)
		n := 0 + rand.Intn(MAX-MIN+1)
		return uint64(n)
	}

	MAX := 1_000_000_000
	STAKED_BAL := genRandUintBelowMax(MAX)
	UNDELEGATED_BAL := genRandUintBelowMax(MAX)
	JUST_DEPOSITED_BAL := genRandUintBelowMax(MAX)

	ST_SUPPLY := genRandUintBelowMax(MAX)

	// s.Require().ElementsMatch([]int{0, 0, 0, 0}, []int{int(STAKED_BAL), int(UNDELEGATED_BAL), int(JUST_DEPOSITED_BAL), int(ST_SUPPLY)}) //
	REDEMPTION_RATE_T0 := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(STAKED_BAL, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	numerator := int64(STAKED_BAL) + int64(UNDELEGATED_BAL) + int64(JUST_DEPOSITED_BAL)
	denominator := int64(ST_SUPPLY)
	expectedNewRate := sdk.NewDec(numerator).Quo(sdk.NewDec(denominator))
	s.Require().Equal(rrNew, expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRatesRandomized_MultipleRuns() {
	for i := 0; i < 100; i++ {
		s.TestUpdateRedemptionRatesRandomized()
		// reset the testing app between runs
		s.Setup()
	}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateZeroStAssets() {

	STAKED_BAL := uint64(5)
	UNDELEGATED_BAL := uint64(3)
	JUST_DEPOSITED_BAL := uint64(3)
	ST_SUPPLY := uint64(0)

	REDEMPTION_RATE_T0 := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(STAKED_BAL, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	// RR should be unchanged
	s.Require().Equal(rrNew, sdk.NewDec(1))
}

func (s *KeeperTestSuite) TestUpdateRedemptionRateZeroNativeAssets() {

	STAKED_BAL := uint64(0)
	UNDELEGATED_BAL := uint64(0)
	JUST_DEPOSITED_BAL := uint64(0)
	ST_SUPPLY := uint64(10)

	REDEMPTION_RATE_T0 := sdk.NewDec(1)
	tc := s.SetupUpdateRedemptionRates(STAKED_BAL, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	// RR should be 0
	s.Require().Equal(rrNew, sdk.NewDec(0))
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoModuleAccountRecords() {

	STAKED_BAL := uint64(5)
	UNDELEGATED_BAL := uint64(3)
	JUST_DEPOSITED_BAL := uint64(3)
	ST_SUPPLY := uint64(10)
	REDEMPTION_RATE_T0 := sdk.NewDec(1)

	tc := s.SetupUpdateRedemptionRates(STAKED_BAL, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, 0)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	numerator := int64(STAKED_BAL) + int64(UNDELEGATED_BAL)
	denominator := int64(ST_SUPPLY)
	expectedNewRate := sdk.NewDec(numerator).Quo(sdk.NewDec(denominator))
	s.Require().Equal(rrNew, expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoStakeDepositRecords() {

	STAKED_BAL := uint64(5)
	UNDELEGATED_BAL := uint64(3)
	JUST_DEPOSITED_BAL := uint64(3)
	ST_SUPPLY := uint64(10)
	REDEMPTION_RATE_T0 := sdk.NewDec(1)

	tc := s.SetupUpdateRedemptionRates(STAKED_BAL, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, 1)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	numerator := int64(STAKED_BAL) + int64(JUST_DEPOSITED_BAL)
	denominator := int64(ST_SUPPLY)
	expectedNewRate := sdk.NewDec(numerator).Quo(sdk.NewDec(denominator))
	s.Require().Equal(rrNew, expectedNewRate)
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoStakedBal() {

	UNDELEGATED_BAL := uint64(3)
	JUST_DEPOSITED_BAL := uint64(3)
	ST_SUPPLY := uint64(10)
	REDEMPTION_RATE_T0 := sdk.NewDec(1)

	// SET HZ STAKED BAL TO 0
	tc := s.SetupUpdateRedemptionRates(0, UNDELEGATED_BAL, JUST_DEPOSITED_BAL, ST_SUPPLY, REDEMPTION_RATE_T0)

	// sanity check on inputs (check redemptionRate at genesis is 1)
	s.Require().Equal(tc.redemption_rate_t0, sdk.NewDec(1))

	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.UpdateRedemptionRates(s.Ctx, records)

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.host_zone.ChainId)
	s.Require().True(found)
	rrNew := hz.RedemptionRate

	numerator := int64(0) + int64(JUST_DEPOSITED_BAL) + int64(UNDELEGATED_BAL)
	denominator := int64(ST_SUPPLY)
	expectedNewRate := sdk.NewDec(numerator).Quo(sdk.NewDec(denominator))
	s.Require().Equal(rrNew, expectedNewRate)
}
