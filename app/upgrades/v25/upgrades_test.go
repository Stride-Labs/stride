package v25_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v24/app/apptesting"
	v25 "github.com/Stride-Labs/stride/v24/app/upgrades/v25"
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
	dummyUpgradeHeight := int64(5)

	// Fund the community pool growth address
	communityGrowthAddress := sdk.MustAccAddressFromBech32(v25.CommunityPoolGrowthAddress)
	s.FundAccount(communityGrowthAddress, sdk.NewCoin(v25.Ustrd, v25.BnocsProposalAmount))

	// Submit upgrade
	s.ConfirmUpgradeSucceededs("v25", dummyUpgradeHeight)

	// Check the transfer was successful
	communityGrowthBalance := s.App.BankKeeper.GetBalance(s.Ctx, communityGrowthAddress, v25.Ustrd)
	s.Require().Zero(communityGrowthBalance.Amount.Int64(), "community growth balance after transfer")

	bnocsCuostidanAddress := sdk.MustAccAddressFromBech32(v25.BnocsCustodian)
	bnocsCustodianBalance := s.App.BankKeeper.GetBalance(s.Ctx, bnocsCuostidanAddress, v25.Ustrd)
	s.Require().Equal(v25.BnocsProposalAmount.Int64(), bnocsCustodianBalance.Amount.Int64(), "bnocs balance")
}
