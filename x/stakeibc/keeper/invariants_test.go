package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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
			res, broken := suite.App.StakeibcKeeper.BalanceStakeHostZoneInvariant()(suite.Ctx)
			suite.Require().Equal(tc.expectedStop, broken, res)
		})
	}
}

func (suite *KeeperTestSuite) TestValidatorWeightHostZoneInvariant() {
	hostZone := types.HostZone{
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
			{
				Name:           "val3",
				Address:        "stride_VAL3",
				CommissionRate: 1,
				Weight:         0,
				Status:         types.Validator_ACTIVE,
				DelegationAmt:  sdk.NewInt(200),
			},
		},
		StakedBal: sdk.NewInt(800),
	}

	testcases := []struct {
		name         string
		hostZone     types.HostZone
		msg          *types.MsgChangeValidatorWeight
		expectedStop bool
	}{
		{
			name:     "happy case",
			hostZone: hostZone,
			msg: &types.MsgChangeValidatorWeight{
				HostZone: hostZone.ChainId,
				ValAddr:  "stride_VAL3",
				Weight:   200,
			},
			expectedStop: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Setup()
			suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, tc.hostZone)
			_, err := suite.GetMsgServer().ChangeValidatorWeight(sdk.WrapSDKContext(suite.Ctx), tc.msg)
			suite.Require().NoError(err)
			res, broken := suite.App.StakeibcKeeper.ValidatorWeightHostZoneInvariant()(suite.Ctx)
			suite.Require().Equal(tc.expectedStop, broken, res)
		})
	}
}

func (suite *KeeperTestSuite) TestRedemptionRateInvariant() {
	
	testcases := []struct {
		name         string
		hostZones    []types.HostZone
		expectedStop bool
	}{
		{
			name: "happy case",
			hostZones: []types.HostZone{
				{
					ChainId:        HostChainId,
					RedemptionRate: sdk.MustNewDecFromStr("1.2"),
				},
				{
					ChainId:        OsmoChainId,
					RedemptionRate: sdk.OneDec(),
				},
			},
			expectedStop: false,
		},
		{
			name: "unhappy case: RedemptionRate is higher than minSafetyThreshold",
			hostZones: []types.HostZone{
				{
					ChainId:        HostChainId,
					RedemptionRate: sdk.MustNewDecFromStr("1.9"),
				},
				{
					ChainId:        OsmoChainId,
					RedemptionRate: sdk.OneDec(),
				},
			},
			expectedStop: true,
		},
		{
			name: "unhappy case: RedemptionRate is smaller than minSafetyThreshold",
			hostZones: []types.HostZone{
				{
					ChainId:        HostChainId,
					RedemptionRate: sdk.MustNewDecFromStr("0.8"),
				},
				{
					ChainId:        OsmoChainId,
					RedemptionRate: sdk.OneDec(),
				},
			},
			expectedStop: true,
		},
	}
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Setup()
			suite.App.StakeibcKeeper.SetParams(suite.Ctx, types.DefaultParams())
			for _, hostZone := range tc.hostZones {
				suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hostZone)
			}

			res, broken := suite.App.StakeibcKeeper.RedemptionRateInvariant()(suite.Ctx)
			suite.Require().Equal(tc.expectedStop, broken, res)
		})
	}
}
