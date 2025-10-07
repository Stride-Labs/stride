package v29_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v29/app/apptesting"
	v29 "github.com/Stride-Labs/stride/v29/app/upgrades/v29"
	strdburnertypes "github.com/Stride-Labs/stride/v29/x/strdburner/types"
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
	// Set an initial total STRD burned
	protocolBurnedAmount := sdkmath.NewInt(1000)
	s.StoreLegacyTotalBurned(protocolBurnedAmount)

	// Run the upgrade
	s.ConfirmUpgradeSucceeded(v29.UpgradeName)

	// Confirm CCV state after upgrade
	consumerParams := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().Equal(consumerParams.ConsumerRedistributionFraction, "1.0")

	// Confirm the burned total state after upgrade
	actualTotal := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	actualProtocol := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	actualUser := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)

	s.Require().Equal(protocolBurnedAmount, actualTotal, "total")
	s.Require().Equal(protocolBurnedAmount, actualProtocol, "protocol")
	s.Require().Equal(sdkmath.ZeroInt(), actualUser, "user")
}

// Helper to write to the legacy total store
func (s *UpgradeTestSuite) StoreLegacyTotalBurned(amount sdkmath.Int) {
	amountBz := sdk.Uint64ToBigEndian(amount.Uint64())
	store := s.Ctx.KVStore(s.App.GetKey(strdburnertypes.StoreKey))
	store.Set([]byte(strdburnertypes.TotalStrdBurnedKey), amountBz)
}
