package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (s *KeeperTestSuite) createNLSMTokenDeposit(n int) []types.LSMTokenDeposit {
	deposits := make([]types.LSMTokenDeposit, n)
	for i := range deposits {
		validatorAddr := "validatorAddress"
		tokenRecordId := strconv.Itoa(i)

		deposits[i].Denom = validatorAddr + tokenRecordId
		deposits[i].ValidatorAddress = validatorAddr
		deposits[i].ChainId = strconv.Itoa(i)
		deposits[i].Amount = sdkmath.NewIntFromUint64(1000)
		deposits[i].Status = types.TRANSFER_IN_PROGRESS
	}
	return deposits
}

func (s *KeeperTestSuite) createSetNLSMTokenDeposit(n int) []types.LSMTokenDeposit {
	newDeposits := s.createNLSMTokenDeposit(n)
	for _, deposit := range newDeposits {
		s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, deposit)
	}
	return newDeposits
}

func (s *KeeperTestSuite) SetGivenLSMTokenDeposit(deposits []types.LSMTokenDeposit) []types.LSMTokenDeposit {
	for _, deposit := range deposits {
		s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, deposit)
	}
	return deposits
}

func (s *KeeperTestSuite) TestLSMTokenDepositGet() {
	deposits := s.createSetNLSMTokenDeposit(10)
	for _, expected := range deposits {
		actual, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().True(found, "deposit not found for chainID %s and denom %s", expected.ChainId, expected.Denom)
		s.Require().Equal(expected, actual, "found deposit did not match expected")
	}
}

func (s *KeeperTestSuite) TestLSMTokenDepositRemove() {
	deposits := s.createSetNLSMTokenDeposit(10)
	for _, expected := range deposits {
		s.App.StakeibcKeeper.RemoveLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		_, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().False(found, "deposit was still found after removal %+v", expected)
	}
}

func (s *KeeperTestSuite) TestLSMTokenDepositGetAll() {
	expected := s.createSetNLSMTokenDeposit(10)
	actual := s.App.StakeibcKeeper.GetAllLSMTokenDeposit(s.Ctx)
	s.Require().Equal(len(actual), len(expected),
		"different number of deposits found %d than was expected %d", len(actual), len(expected))
	s.Require().ElementsMatch(actual, expected, "actual list did not match expected list")
}

func (s *KeeperTestSuite) TestLSMTokenDepositAdd() {
	// existing will have the same chainId, denoms, and recordIds as the first 5 in new
	// they are deposits of the same "type" in each parallel index, last 5 in new are net new
	existingDeposits := s.createSetNLSMTokenDeposit(5)
	newDeposits := s.createNLSMTokenDeposit(10)

	// verify non-overlapping half of the newDeposits do not yet exist in the store
	for i := 5; i < 10; i++ {
		new := newDeposits[i]
		_, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, new.ChainId, new.Denom)
		s.Require().False(found, "deposit was unexpectedly found in store already %+v", new)
	}

	// call Add on all newDeposits so they are in store one way or another
	for _, deposit := range newDeposits {
		s.App.StakeibcKeeper.AddLSMTokenDeposit(s.Ctx, deposit)
	}

	// verify that for previously existing deposits the amounts add
	for i := 0; i < 5; i++ {
		existing := existingDeposits[i]
		new := newDeposits[i]
		actual, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, new.ChainId, new.Denom)
		s.Require().True(found, "deposit not found in store %+v", new)
		s.Require().Equal(actual.Amount, sdkmath.Int.Add(existing.Amount, new.Amount),
			"found amount %d did not match expected sum %d + %d = %d", actual.Amount,
			existing.Amount, new.Amount, sdkmath.Int.Add(existing.Amount, new.Amount))
	}

	// verify that for previously non-existing deposits the amounts set
	for i := 5; i < 10; i++ {
		new := newDeposits[i]
		actual, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, new.ChainId, new.Denom)
		s.Require().True(found, "deposit not found in store %+v", new)
		s.Require().Equal(actual.Amount, new.Amount,
			"found amount %d did not match expected amount %d ", actual.Amount, new.Amount)
	}
}

func (s *KeeperTestSuite) TestLSMTokenDepositStatusUpdate() {
	statuses := []types.LSMDepositStatus{
		types.TRANSFER_IN_PROGRESS,
		types.TRANSFER_FAILED,
		types.DETOKENIZATION_QUEUE,
		types.DETOKENIZATION_IN_PROGRESS,
		types.DETOKENIZATION_FAILED,
	}
	deposits := s.createSetNLSMTokenDeposit(5)
	for i, status := range statuses {
		s.App.StakeibcKeeper.UpdateLSMTokenDepositStatus(s.Ctx, deposits[i], status)
	}

	for i, deposit := range deposits {
		actual, _ := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, deposit.ChainId, deposit.Denom)
		s.Require().Equal(actual.Status, statuses[i], "status did not update for example %d", i)
	}
}

func (s *KeeperTestSuite) TestLSMDepositsForHostZoneGet() {
	// For HostZone with id i there will be i+1 deposits created, all denom unique
	deposits := s.createNLSMTokenDeposit(15) // 15 = 1 + 2 + 3 + 4 + 5
	idx := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < i+1; j++ {
			deposits[idx].ChainId = strconv.Itoa(i)
			idx++
		}
	}
	s.SetGivenLSMTokenDeposit(deposits)

	// Check there are i+1 deposits for chainid i, all deposits returned are from right chain
	for i := 0; i < 5; i++ {
		hostChainId := strconv.Itoa(i)
		chainDeposits := s.App.StakeibcKeeper.GetLSMDepositsForHostZone(s.Ctx, hostChainId)
		s.Require().Equal(len(chainDeposits), i+1, "Unexpected number of deposits found for chainId %d", i)
		for _, deposit := range chainDeposits {
			s.Require().Equal(deposit.ChainId, hostChainId, "Got a deposit from the wrong chain!")
		}
	}
}

func (s *KeeperTestSuite) TestLSMDepositsForHostZoneWithStatusGet() {
	// Check that we get every deposit which matches hostzone and status
	// Check that we only get deposits which match hostzone and status
	// Need a predictable, different, non-zero number of deposits for each (zone,status) combo
	// Deposit status is actually a 0 indexed int32 so make (hostzone+1)*(status+1) deposits
	statuses := []types.LSMDepositStatus{
		types.TRANSFER_IN_PROGRESS,
		types.TRANSFER_FAILED,
		types.DETOKENIZATION_QUEUE,
		types.DETOKENIZATION_IN_PROGRESS,
		types.DETOKENIZATION_FAILED,
	}
	deposits := s.createNLSMTokenDeposit(225) // total number across all combos
	idx := 0
	for hzid := 0; hzid < 5; hzid++ {
		for sid := 0; sid < len(statuses); sid++ {
			numCombo := (hzid + 1) * (sid + 1)
			for i := 0; i < numCombo; i++ {
				deposits[idx].ChainId = strconv.Itoa(hzid)
				deposits[idx].Status = statuses[sid]
				idx++
			}
			s.Ctx.Logger().Info("combonum %d idx %d", numCombo, idx)
		}
	}
	s.SetGivenLSMTokenDeposit(deposits)

	for hzid := 0; hzid < 5; hzid++ {
		for sid := 0; sid < len(statuses); sid++ {
			expectedLen := (hzid + 1) * (sid + 1)
			chainId := strconv.Itoa(hzid)
			status := statuses[sid]
			actual := s.App.StakeibcKeeper.GetLSMDepositsForHostZoneWithStatus(s.Ctx, chainId, status)
			// Check that we get every deposit which matches hostzone and status
			s.Require().Equal(len(actual), expectedLen, "Unexpected number of deposits found for chainId %d", hzid)
			// Check that we only get deposits which match hostzone and status
			for _, deposit := range actual {
				s.Require().Equal(deposit.ChainId, chainId, "Got back deposit from different chain!")
				s.Require().Equal(deposit.Status, status, "Got back deposit with wrong status!")
			}
		}
	}
}
