package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/records/types"
)

func (s *KeeperTestSuite) createNLSMTokenDeposit(n int) []types.LSMTokenDeposit {
	deposits := make([]types.LSMTokenDeposit, n)
	for i := range deposits {
		validatorAddr := "validatorAddress"
		tokenRecordId := strconv.Itoa(i)

		deposits[i].Denom = validatorAddr + tokenRecordId
		deposits[i].IbcDenom = "ibc/" + validatorAddr + tokenRecordId
		deposits[i].ValidatorAddress = validatorAddr
		deposits[i].ChainId = strconv.Itoa(i)
		deposits[i].Amount = sdkmath.NewIntFromUint64(1000)
		deposits[i].Status = types.LSMTokenDeposit_DEPOSIT_PENDING
		deposits[i].StToken = sdk.NewCoin("sttoken", sdk.NewInt(int64(i)))
	}
	return deposits
}

func (s *KeeperTestSuite) setGivenLSMTokenDeposit(deposits []types.LSMTokenDeposit) []types.LSMTokenDeposit {
	for _, deposit := range deposits {
		s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)
	}
	return deposits
}

func (s *KeeperTestSuite) createSetNLSMTokenDeposit(n int) []types.LSMTokenDeposit {
	newDeposits := s.createNLSMTokenDeposit(n)
	return s.setGivenLSMTokenDeposit(newDeposits)
}

func (s *KeeperTestSuite) TestGetLSMTokenDeposit() {
	deposits := s.createSetNLSMTokenDeposit(10)
	for _, expected := range deposits {
		actual, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().True(found, "deposit not found for chainID %s and denom %s", expected.ChainId, expected.Denom)
		s.Require().Equal(expected, actual, "found deposit did not match expected")
	}
}

func (s *KeeperTestSuite) TestRemoveLSMTokenDeposit() {
	deposits := s.createSetNLSMTokenDeposit(10)
	for _, expected := range deposits {
		s.App.RecordsKeeper.RemoveLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		_, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().False(found, "deposit was still found after removal %+v", expected)
	}
}

func (s *KeeperTestSuite) TestGetAllLSMTokenDeposit() {
	expected := s.createSetNLSMTokenDeposit(10)
	actual := s.App.RecordsKeeper.GetAllLSMTokenDeposit(s.Ctx)
	s.Require().Equal(len(expected), len(actual),
		"different number of deposits found %d than was expected %d", len(actual), len(expected))
	s.Require().ElementsMatch(actual, expected, "actual list did not match expected list")
}

func (s *KeeperTestSuite) TestUpdateLSMTokenDepositStatus() {
	statuses := []types.LSMTokenDeposit_Status{
		types.LSMTokenDeposit_DEPOSIT_PENDING,
		types.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		types.LSMTokenDeposit_TRANSFER_FAILED,
		types.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		types.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
		types.LSMTokenDeposit_DETOKENIZATION_FAILED,
	}
	deposits := s.createSetNLSMTokenDeposit(len(statuses))
	for i, status := range statuses {
		s.App.RecordsKeeper.UpdateLSMTokenDepositStatus(s.Ctx, deposits[i], status)
	}

	for i, deposit := range deposits {
		actual, _ := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, deposit.ChainId, deposit.Denom)
		s.Require().Equal(actual.Status, statuses[i], "status did not update for example %d", i)
	}
}

func (s *KeeperTestSuite) TestGetLSMDepositsForHostZone() {
	// For HostZone with id i there will be i+1 deposits created, all denom unique
	// i.e. {chain-0, chain-1, chain-1, chain-2, chain-2, chain-2, ...}
	deposits := s.createNLSMTokenDeposit(15) // 15 = 1 + 2 + 3 + 4 + 5
	idx := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < i+1; j++ {
			deposits[idx].ChainId = strconv.Itoa(i)
			idx++
		}
	}
	s.setGivenLSMTokenDeposit(deposits)

	// Check there are i+1 deposits for chainid i, all deposits returned are from right chain
	for i := 0; i < 5; i++ {
		hostChainId := strconv.Itoa(i)
		chainDeposits := s.App.RecordsKeeper.GetLSMDepositsForHostZone(s.Ctx, hostChainId)
		s.Require().Equal(i+1, len(chainDeposits), "Unexpected number of deposits found for chainId %d", i)
		for _, deposit := range chainDeposits {
			s.Require().Equal(hostChainId, deposit.ChainId, "Got a deposit from the wrong chain!")
		}
	}
}

func (s *KeeperTestSuite) TestGetLSMDepositsForHostZoneWithStatus() {
	// Necessary to check that we get every deposit for a given hostzone and status
	// Necessary to also check that we *only* get deposits which match hostzone and status
	// Need a predictable, different, non-zero number of deposits for each (zone, status) combo
	numHostZones := 5
	statuses := []types.LSMTokenDeposit_Status{
		types.LSMTokenDeposit_DEPOSIT_PENDING,
		types.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		types.LSMTokenDeposit_TRANSFER_FAILED,
		types.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		types.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
		types.LSMTokenDeposit_DETOKENIZATION_FAILED,
	}

	// For each (zone, status) combo, create a different number of deposits
	//  determined by numDeposits = (hostZone index + 1) * (status index + 1)

	// For instance:
	//    chain-0, status-0 => 1 deposit
	//    chain-0, status-1 => 1 * 2 = 2 deposit
	//    chain-1, status-1 => 2 * 2 = 4 deposits

	deposits := s.createNLSMTokenDeposit(315) // 315 = 15 * 21 is total number across all combos
	// nZones = 5 --> 15 = 1 + 2 + 3 + 4 + 5      nStatuses = 6 -->  21 = 1 + 2 + 3 + 4 + 5 + 6
	// Generally with nZones number of host zones and nStatuses number of statuses
	//   there will be a totalDeposits = 1/4 * nZones * (nZones + 1) * nStatuses * (nStatuses + 1)

	idx := 0
	for hzid := 0; hzid < numHostZones; hzid++ {
		for sid := 0; sid < len(statuses); sid++ {
			numCombo := (hzid + 1) * (sid + 1)
			for i := 0; i < numCombo; i++ {
				deposits[idx].ChainId = strconv.Itoa(hzid)
				deposits[idx].Status = statuses[sid]
				idx++
			}
		}
	}
	s.setGivenLSMTokenDeposit(deposits)

	for hzid := 0; hzid < numHostZones; hzid++ {
		for sid := 0; sid < len(statuses); sid++ {
			expectedLen := (hzid + 1) * (sid + 1)
			chainId := strconv.Itoa(hzid)
			status := statuses[sid]
			actual := s.App.RecordsKeeper.GetLSMDepositsForHostZoneWithStatus(s.Ctx, chainId, status)
			// Check that we get every deposit which matches hostzone and status
			s.Require().Equal(expectedLen, len(actual), "Unexpected number of deposits found for chainId %d", hzid)
			// Check that we only get deposits which match hostzone and status
			for _, deposit := range actual {
				s.Require().Equal(chainId, deposit.ChainId, "Got back deposit from different chain!")
				s.Require().Equal(status, deposit.Status, "Got back deposit with wrong status!")
			}
		}
	}
}
