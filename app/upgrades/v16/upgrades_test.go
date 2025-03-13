package v16_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

var CosmosHubChainIdTest = "cosmoshub-4"

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	// Setup the store before the upgrade
	checkCosmosHubAfterUpgrade := s.SetupHostZonesBeforeUpgrade()

	// Run the upgrade to set the bounds and clear pending queries
	s.ConfirmUpgradeSucceededs("v16", dummyUpgradeHeight)

	// Check the store after the upgrade
	checkCosmosHubAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupHostZonesBeforeUpgrade() func() {
	// Create 10 dummy host zones
	for i := 0; i < 10; i++ {
		chainId := fmt.Sprintf("chain-%d", i)

		hostZone := stakeibctypes.HostZone{
			ChainId:                chainId,
			Halted:                 false,
			RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("1.0"),
			MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"),
			MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.97"),
			MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.05"),
			MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.10"),
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}
	// create Cosmos Hub Host Zone
	hostZone := stakeibctypes.HostZone{
		ChainId:                CosmosHubChainIdTest,
		Halted:                 true,
		RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("1.0"),
		MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"),
		MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.97"),
		MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.05"),
		MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.10"),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return func() {
		hostZones := s.App.StakeibcKeeper.GetAllHostZone(s.Ctx)

		for _, hostZone := range hostZones {
			s.Require().False(hostZone.Halted, fmt.Sprintf("host zone %s should not be halted: %v", hostZone.ChainId, hostZone))
		}
		// Confirm Cosmos Hub host zone is not unhalted
		cosmosHubHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, CosmosHubChainIdTest)
		s.Require().True(found, "Cosmos Hub host zone not found!")
		s.Require().False(cosmosHubHostZone.Halted, "Cosmos Hub host zone should not be halted")
	}
}
