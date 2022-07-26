package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupStakeExistingDepositsOnHostZones() {
	initialDepositAmount := int64(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(ibcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(stAtom, 0),
	}
	s.FundAccount(user.acc, user.atomBalance)

	module := Account{
		acc:           s.App.AccountKeeper.GetModuleAddress(stakeibc.ModuleName),
		atomBalance:   sdk.NewInt64Coin(ibcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(stAtom, 10_000_000),
	}
	s.FundModuleAccount(stakeibc.ModuleName, module.atomBalance)
	s.FundModuleAccount(stakeibc.ModuleName, module.stAtomBalance)

	hostZone := stakeibc.HostZone{
		ChainId:        "GAIA",
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             initialDepositAmount,
	}

	// append 10 deposit records to the deposit record list
	for i := 0; i < 10; i++ {
		s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, initialDepositRecord)
	}
	return
}

func (s *KeeperTestSuite) TestStakeExistingDepositsOnHostZonesSuccessful() {
	s.SetupStakeExistingDepositsOnHostZones()

	PrevDepositRecordCount := s.App.StakeibcKeeper.RecordsKeeper.GetDepositRecordCount(s.Ctx)
	depositRecords := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.App.StakeibcKeeper.StakeExistingDepositsOnHostZones(s.Ctx, 100, depositRecords)
	NewDepositRecordCount := s.App.StakeibcKeeper.RecordsKeeper.GetDepositRecordCount(s.Ctx)

	expectedNumProcessedWithCap := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibc.KeyMaxICACallsPerEpoch)
	s.Require().Equal(int64(PrevDepositRecordCount-expectedNumProcessedWithCap), int64(NewDepositRecordCount))
}
