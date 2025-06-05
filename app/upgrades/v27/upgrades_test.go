package v27_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	v27 "github.com/Stride-Labs/stride/v27/app/upgrades/v27"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
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

	// Create gaia host zone with lsm disabled
	gaiaHostZone := stakeibctypes.HostZone{
		ChainId:               v27.GaiaChainId,
		LsmLiquidStakeEnabled: false,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, gaiaHostZone)

	// Create evmos host zone with delegation changes in progress
	evmosHostZone := stakeibctypes.HostZone{
		ChainId: v27.EvmosChainId,
		Validators: []*stakeibctypes.Validator{
			{Name: "val1", DelegationChangesInProgress: 0},
			{Name: "val2", DelegationChangesInProgress: 1},
			{Name: "val3", DelegationChangesInProgress: 2},
			{Name: "val4", DelegationChangesInProgress: 3},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, evmosHostZone)

	// Run upgrade
	s.ConfirmUpgradeSucceededs(v27.UpgradeName, upgradeHeight)

	// Confirm LSM was enabled
	gaiaHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v27.GaiaChainId)
	s.Require().True(found, "gaia found")
	s.Require().True(gaiaHostZone.LsmLiquidStakeEnabled, "lsm enabled")

	// Confirm delegation changes
	expectedDelegationChanges := map[string]int64{
		"val1": 0,
		"val2": 0,
		"val3": 1,
		"val4": 2,
	}
	evmosHostZone, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, v27.EvmosChainId)
	s.Require().True(found, "evmos found")

	for _, validator := range evmosHostZone.Validators {
		s.Require().Equal(expectedDelegationChanges[validator.Name],
			validator.DelegationChangesInProgress, "%s delegation change", validator.Name)
	}
}
