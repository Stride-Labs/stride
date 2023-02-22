package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func (suite *KeeperTestSuite) TestBalanceStakeHostZoneInvariant() {
	testcases := []struct {
		name         string
		hostZone     types.HostZone
		expectedStop bool
	}{
		{
			name: "unhappy case",
			hostZone: types.HostZone{
				ChainId: HostChainId,
				Validators: []*types.Validator{
					{
						Name:           "val1",
						Address:        "stride_VAL1",
						CommissionRate: 1,
						Weight:         100,
						Status:         types.Validator_ACTIVE,
						DelegationAmt:  sdk.NewInt(150),
					},
					{
						Name:           "val2",
						Address:        "stride_VAL2",
						CommissionRate: 2,
						Weight:         500,
						Status:         types.Validator_ACTIVE,
						DelegationAmt:  sdk.NewInt(500),
					},
				},
				StakedBal: sdk.NewInt(600),
			},
			expectedStop: true,
		},
		{
			name: "happy case",
			hostZone: types.HostZone{
				ChainId: HostChainId,
				Validators: []*types.Validator{
					{
						Name:           "val1",
						Address:        "stride_VAL1",
						CommissionRate: 1,
						Weight:         100,
						Status:         types.Validator_ACTIVE,
						DelegationAmt:  sdk.NewInt(100),
					},
					{
						Name:           "val2",
						Address:        "stride_VAL2",
						CommissionRate: 2,
						Weight:         500,
						Status:         types.Validator_ACTIVE,
						DelegationAmt:  sdk.NewInt(500),
					},
				},
				StakedBal: sdk.NewInt(600),
			},
			expectedStop: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Setup()
			suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, tc.hostZone)
			res, broken := suite.App.StakeibcKeeper.DelegationsSumToStakedBal()(suite.Ctx)
			suite.Require().Equal(tc.expectedStop, broken, res)
		})
	}
}
