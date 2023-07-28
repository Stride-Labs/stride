package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	stakeibctypes "github.com/Stride-Labs/stride/v12/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestRegisterStTokenDenomsToWhitelist() {
	tc := s.SetupRegisterHostZone()
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "able to successfully register host zone")

	for _, tc := range []struct {
		desc              string
		newDenoms         []string
		expectedWhitelist []string
		expectedErr       error
	}{
		{
			desc:              "both valid and invalid denoms",
			newDenoms:         []string{"stuatom", "stuosmo"},
			expectedWhitelist: nil,
			expectedErr:       stakeibctypes.ErrStTokenNotFound,
		},
		{
			desc:              "only invalid denoms",
			newDenoms:         []string{"stuosmo", "stujuno"},
			expectedWhitelist: nil,
			expectedErr:       stakeibctypes.ErrStTokenNotFound,
		},
		{
			desc:              "only valid denoms",
			newDenoms:         []string{"stuatom"},
			expectedWhitelist: []string{"stuatom"},
			expectedErr:       nil,
		},
		{
			desc:              "empty",
			newDenoms:         []string{},
			expectedWhitelist: []string{"stuatom"},
			expectedErr:       nil,
		},
	} {
		s.Run(tc.desc, func() {
			err := s.App.StakeibcKeeper.RegisterStTokenDenomsToWhitelist(s.Ctx, tc.newDenoms)
			s.Require().Equal(tc.expectedErr, err)

			params := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
			s.Require().Equal(tc.expectedWhitelist, params.RewardDenoms)
		})
	}
}
