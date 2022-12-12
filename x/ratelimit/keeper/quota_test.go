package keeper_test

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func (s *KeeperTestSuite) createQuotas() []types.Quota {
	quotas := []types.Quota{}
	for i := 1; i <= 5; i++ {
		quota := types.Quota{
			Name:            fmt.Sprintf("name-%d", i),
			MaxPercentRecv:  uint64(i),
			MaxPercentSend:  uint64(i),
			DurationMinutes: uint64(i),
		}

		quotas = append(quotas, quota)
		s.App.RatelimitKeeper.SetQuota(s.Ctx, quota)
	}
	return quotas
}

func (s *KeeperTestSuite) TestGetQuota() {
	quotas := s.createQuotas()
	expectedQuota := quotas[0]

	actualQuota, found := s.App.RatelimitKeeper.GetQuota(s.Ctx, "name-1")
	s.Require().True(found, "element should have been found, but was not")
	s.Require().Equal(expectedQuota, actualQuota)
}

func (s *KeeperTestSuite) TestRemoveQuota() {
	quotas := s.createQuotas()
	idToRemove := quotas[0].Name

	s.App.RatelimitKeeper.RemoveQuota(s.Ctx, idToRemove)
	_, found := s.App.RatelimitKeeper.GetQuota(s.Ctx, idToRemove)
	s.Require().False(found, "element should have been removed, but was found")
}

func (s *KeeperTestSuite) TestGetAllQuotas() {
	expectedQuotas := s.createQuotas()
	actualQuotas := s.App.RatelimitKeeper.GetAllQuotas(s.Ctx)
	s.Require().Len(actualQuotas, len(expectedQuotas))
	s.Require().ElementsMatch(expectedQuotas, actualQuotas, "all quotas")
}
