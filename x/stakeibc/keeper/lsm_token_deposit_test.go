package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (s *KeeperTestSuite) createSetNLSMTokenDeposit(ctx sdk.Context, n int) []types.LSMTokenDeposit {
	items := make([]types.LSMTokenDeposit, n)
	for i := range items {
		validatorAddr := "validatorAddress"
		tokenRecordId := strconv.Itoa(i)

		items[i].Denom = validatorAddr + tokenRecordId
		items[i].ValidatorAddress = validatorAddr
		items[i].ChainId = strconv.Itoa(i)
		items[i].Amount = sdkmath.ZeroInt()
		items[i].Status = types.TRANSFER_IN_PROGRESS
		s.App.StakeibcKeeper.SetLSMTokenDeposit(ctx, items[i])
	}
	return items
}

func (s *KeeperTestSuite) TestLSMTokenDepositGet() {
	deposits := s.createSetNLSMTokenDeposit(s.Ctx, 10)
	for _, expected := range deposits {
		actual, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, expected.ChainId, expected.Denom)
		s.Require().True(found, "deposit not found for chainID %s and denom %s", expected.ChainId, expected.Denom)
		s.Require().Equal(expected, actual, "found deposit %+v did not match expected %+v", actual, expected)
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
	s.Require().ElementsMatch(actual, expected, "actual %+v was not expected %+v", actual, expected)
}
