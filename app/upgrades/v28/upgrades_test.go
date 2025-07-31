package v28_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	v28 "github.com/Stride-Labs/stride/v27/app/upgrades/v28"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinOuterRedemptionRate sdk.Dec
	ExpectedMaxOuterRedemptionRate sdk.Dec
}

type UpdateRedemptionRateInnerBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinInnerRedemptionRate sdk.Dec
	ExpectedMaxInnerRedemptionRate sdk.Dec
}

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

	// ------- Set state before upgrade -------
	checkRedemptionRates := s.SetupTestUpdateRedemptionRateBounds()
	checkLockedTokens := s.SetupTestDeliverLockedTokens()

	// ------- Run upgrade -------
	s.ConfirmUpgradeSucceededs(v28.UpgradeName, upgradeHeight)

	// ------- Confirm state after upgrade -------
	checkRedemptionRates()
	checkLockedTokens()
}

func (s *UpgradeTestSuite) CreateDelayedVestingAccount(address sdk.AccAddress, end int64, coins int64) *types.DelayedVestingAccount {
	afterStartTime := time.Unix(1753991194, 0) // late August 2025
	endTime := time.Unix(end, 0)

	// init a base account
	bacc := authtypes.NewBaseAccountWithAddress(address)

	// send locked tokens to the base account
	origCoins := sdk.Coins{sdk.NewInt64Coin(s.App.StakingKeeper.BondDenom(s.Ctx), coins)}
	dva := vesting.NewDelayedVestingAccount(bacc, origCoins, end)

	// require no coins vested shortly after the start of the vesting schedule)
	vestedCoins := dva.GetVestedCoins(afterStartTime)
	s.Require().Nil(vestedCoins)

	// require all coins vested at the end of the vesting schedule)
	vestedCoins = dva.GetVestedCoins(endTime)
	s.Require().Equal(origCoins, vestedCoins)

	return dva
}

func (s *UpgradeTestSuite) SetupTestDeliverLockedTokens() func() {

	// Init DelayedVestingAccount
	deliveryAccountAddress, err := sdk.AccAddressFromBech32(v28.DeliveryAccount)
	s.Require().NoError(err)
	deliveryAccount := s.CreateDelayedVestingAccount(deliveryAccountAddress, v28.VestingEndTime, v28.LockedTokenAmount)
	// Also needs to be added to the account keeper
	s.App.AccountKeeper.SetAccount(s.Ctx, deliveryAccount)

	// Fund account and test sending a tx to mimic mainnet
	s.FundAccount(deliveryAccountAddress, sdk.NewCoin(s.App.StakingKeeper.BondDenom(s.Ctx), sdkmath.NewInt(1_000_000)))
	// Account sends some unlocked tokens
	s.App.BankKeeper.SendCoins(s.Ctx, deliveryAccountAddress, deliveryAccountAddress, sdk.NewCoins(sdk.NewCoin(s.App.StakingKeeper.BondDenom(s.Ctx), sdkmath.NewInt(500_000))))

	// Return callback to check store after upgrade
	return func() {
		account := s.App.AccountKeeper.GetAccount(s.Ctx, sdk.MustAccAddressFromBech32(v28.DeliveryAccount))

		// Check that the account is a DelayedVestingAccount
		delayedVestingAccount, ok := account.(*vesting.DelayedVestingAccount)
		s.Require().True(ok)

		// Check that the end time is set correctly
		s.Require().Equal(v28.VestingEndTime, delayedVestingAccount.EndTime)

		// Check that the original vesting amount is set correctly
		expectedAmount := sdk.NewCoins(sdk.NewCoin("ustrd", sdk.NewInt(v28.LockedTokenAmount)))
		s.Require().Equal(expectedAmount, delayedVestingAccount.OriginalVesting)
	}
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.5"), // 1 - 50% = 0.95
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("2.0"), // 1 + 100% = 1.25
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.55"), // 1.1 - 50% = 0.55
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("2.2"),  // 1.1 + 100% = 2.2
		},
	}

	// Create a host zone for each test case
	for _, tc := range testCases {
		hostZone := stakeibctypes.HostZone{
			ChainId:        tc.ChainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Return callback to check store after upgrade
	return func() {
		// Confirm they were all updated
		for _, tc := range testCases {
			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.ChainId)
			s.Require().True(found)

			s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "%s - min outer", tc.ChainId)
			s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "%s - max outer", tc.ChainId)
		}
	}
}
