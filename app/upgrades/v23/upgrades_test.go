package v23_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v25/app/apptesting"
	v23 "github.com/Stride-Labs/stride/v25/app/upgrades/v23"
	recordstypes "github.com/Stride-Labs/stride/v25/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v25/x/stakeibc/types"
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
	dummyUpgradeHeight := int64(4)

	minTransferAmount := sdkmath.NewInt(100)
	initialDetokenizeAmount := sdkmath.NewInt(100)
	expectedDetokenizeAmount := sdkmath.NewInt(99)

	// Set the allowed ibc clients to an empty list
	s.App.IBCKeeper.ClientKeeper.SetParams(s.Ctx, ibcclienttypes.Params{AllowedClients: []string{}})

	// Create a trade route with the deprecated trade config
	tradeRoutes := stakeibctypes.TradeRoute{
		HostDenomOnHostZone:     "host-denom",
		RewardDenomOnRewardZone: "reward-denom",

		TradeConfig: stakeibctypes.TradeConfig{ //nolint:staticcheck
			MinSwapAmount: minTransferAmount,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoutes)

	// Create the failed detokenization record
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordstypes.LSMTokenDeposit{
		ChainId: v23.CosmosChainId,
		Denom:   v23.FailedLSMDepositDenom,
		Amount:  initialDetokenizeAmount,
	})

	// Run the upgrade
	s.ConfirmUpgradeSucceededs(v23.UpgradeName, dummyUpgradeHeight)

	// Confirm trade route was migrated
	for _, tradeRoute := range s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx) {
		s.Require().Equal(tradeRoute.MinTransferAmount, minTransferAmount)
	}

	// Confirm the ibc wasm client was added
	params := s.App.IBCKeeper.ClientKeeper.GetParams(s.Ctx)
	s.Require().Equal([]string{ibcwasmtypes.Wasm}, params.AllowedClients, "ibc allowed clients")

	// Confirm the lsm deposit record was reset
	lsmRecord, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v23.CosmosChainId, v23.FailedLSMDepositDenom)
	s.Require().True(found, "lsm deposit record should have been found")
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE, lsmRecord.Status, "lsm record status")
	s.Require().Equal(expectedDetokenizeAmount, lsmRecord.Amount, "lsm deposit record amount")
}
