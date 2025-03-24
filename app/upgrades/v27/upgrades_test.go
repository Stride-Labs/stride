package v27_test

import (
	"testing"

	"github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	v27 "github.com/Stride-Labs/stride/v26/app/upgrades/v27"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	upgradeHeight := int64(4)

	s.ConfirmUpgradeSucceededs(v27.UpgradeName, upgradeHeight)

	// Confirm consumer ID is set to 1
	params := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().Equal(params.ConsumerId, "1")
}

func (s *UpgradeTestSuite) TestDistributionFix() {
	jsonDistGenesis := os.MustReadFile("test_dist_genesis.json")
	jsonStakingGenesis := os.MustReadFile("test_staking_genesis.json")

	// Load faulty state from json
	var distGenesisState disttypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonDistGenesis, &distGenesisState)
	var stakingGenesisState stakingtypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonStakingGenesis, &stakingGenesisState)

	// Align x/bank modules with faulty state
	for i := range distGenesisState.OutstandingRewards {
		coins, _ := distGenesisState.OutstandingRewards[i].OutstandingRewards.TruncateDecimal()
		for _, coin := range coins {
			s.FundModuleAccount(disttypes.ModuleName, coin)
		}
	}
	s.FundModuleAccount(stakingtypes.BondedPoolName, types.NewInt64Coin("stake", 1038549945))
	s.FundModuleAccount(stakingtypes.NotBondedPoolName, types.NewInt64Coin("stake", 220000))

	// Overwrite x/staking's state with imported
	s.App.StakingKeeper.InitGenesis(s.Ctx, &stakingGenesisState)

	// Overwrite x/distribution's state with faulty state
	s.App.DistrKeeper.InitGenesis(s.Ctx, distGenesisState)

	// Get validator address
	valAddrResp, err := stakingkeeper.NewQuerier(&s.App.StakingKeeper).Validators(s.Ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.Bonded.String(),
	})
	s.Require().NoError(err)

	valAddr, err := types.ValAddressFromBech32(valAddrResp.Validators[0].OperatorAddress)
	s.Require().NoError(err)

	// Verify that things are failing
	for _, dsi := range distGenesisState.DelegatorStartingInfos {
		delAddr := types.MustAccAddressFromBech32(dsi.DelegatorAddress)

		_, err = s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, delAddr, valAddr)
		s.Require().Error(err)
		s.Require().ErrorContains(err, "asdasdasd")
	}

	// TODO Fix x/ditribution state

	// TODO Verify that things are working
}
