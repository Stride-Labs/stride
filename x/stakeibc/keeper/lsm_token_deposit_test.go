package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (s *KeeperTestSuite) createNLSMTokenDeposit(ctx sdk.Context, n int) []types.LSMTokenDeposit {
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

func (s *KeeperTestSuite) createSetNLSMTokenDeposit(ctx sdk.Context, n int) []types.LSMTokenDeposit {
	newDeposits := s.createNLSMTokenDeposit(s.Ctx, n)
	for _, deposit := range newDeposits {
		s.App.StakeibcKeeper.SetLSMTokenDeposit(ctx, deposit)
	}
	return newDeposits
}

func (s *KeeperTestSuite) TestLSMTokenDepositGet() {
	deposits := s.createSetNLSMTokenDeposit(s.Ctx, 10)
	for _, expected := range deposits {
		actual, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().True(found, "deposit not found for chainID %s and denom %s", expected.ChainId, expected.Denom)
		s.Require().Equal(expected, actual, "found deposit did not match expected")
	}
}

func (s *KeeperTestSuite) TestLSMTokenDepositRemove() {
	deposits := s.createSetNLSMTokenDeposit(s.Ctx, 10)
	for _, expected := range deposits {
		s.App.StakeibcKeeper.RemoveLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		_, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().False(found, "deposit was still found after removal %+v", expected)
	}
}

func (s *KeeperTestSuite) TestLSMTokenDepositGetAll() {
	expected := s.createSetNLSMTokenDeposit(s.Ctx, 10)
	actual := s.App.StakeibcKeeper.GetAllLSMTokenDeposit(s.Ctx)
	s.Require().Equal(len(actual), len(expected),
		"different number of deposits found %d than was expected %d", len(actual), len(expected))
	s.Require().ElementsMatch(actual, expected, "actual list did not match expected list")
}

func (s *KeeperTestSuite) TestLSMTokenDepositAdd() {
	// existing will have the same chainId, denoms, and recordIds as the first 5 in new
	// they are deposits of the same "type" in each parallel index, last 5 in new are net new
	existingDeposits := s.createSetNLSMTokenDeposit(s.Ctx, 5)
	newDeposits := s.createNLSMTokenDeposit(s.Ctx, 10)

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
