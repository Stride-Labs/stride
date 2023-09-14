package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestRegisterStTokenDenomsToWhitelist() {
	tc := s.SetupRegisterHostZone()
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "able to successfully register host zone")

	// RegisterHostZone should have already registered stToken to consumer reward denom whitelist
	params := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	stDenom := stakeibctypes.StAssetDenomFromHostZoneDenom(tc.validMsg.HostDenom)
	expectedWhitelist := []string{stDenom}
	s.Require().Equal([]string{stDenom}, params.RewardDenoms)

	for _, tc := range []struct {
		desc              string
		newDenoms         []string
		expectedWhitelist []string
		expectedErr       error
	}{
		{
			desc:              "both valid and invalid denoms",
			newDenoms:         []string{stDenom, "stuosmo"},
			expectedWhitelist: expectedWhitelist,
			expectedErr:       stakeibctypes.ErrStTokenNotFound,
		},
		{
			desc:              "only invalid denoms",
			newDenoms:         []string{"stuosmo", "stujuno"},
			expectedWhitelist: expectedWhitelist,
			expectedErr:       stakeibctypes.ErrStTokenNotFound,
		},
		{
			desc:              "only valid denoms",
			newDenoms:         []string{stDenom},
			expectedWhitelist: expectedWhitelist,
			expectedErr:       nil,
		},
		{
			desc:              "empty",
			newDenoms:         []string{},
			expectedWhitelist: expectedWhitelist,
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
