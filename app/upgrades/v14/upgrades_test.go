package v14_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v13/app/apptesting"
	v14 "github.com/Stride-Labs/stride/v13/app/upgrades/v14"
)

var (
	stakeDenom         = "ustrd"
	emptyCoins         = sdk.Coins{}
	dummyUpgradeHeight = int64(5)
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
	s.SetupVestingStoreBeforeUpgrade()
	s.ConfirmUpgradeSucceededs("v14", dummyUpgradeHeight)
	s.CheckVestingStoreAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupVestingStoreBeforeUpgrade() {
	// Initialize the two accounts as continuous vesting accounts
	// Create the ContinuousVestingAccount
	address1, err := sdk.AccAddressFromBech32(v14.Account1)
	s.Require().NoError(err)
	address2, err := sdk.AccAddressFromBech32(v14.Account2)
	s.Require().NoError(err)
	account1 := s.CreateContinuousVestingAccount(address1, v14.VestingStartTimeAccount1, v14.VestingEndTimeAccount1, v14.Account1VestingUstrd)
	account2 := s.CreateContinuousVestingAccount(address2, v14.VestingStartTimeAccount2, v14.VestingEndTimeAccount2, v14.Account2VestingUstrd)

	// Fund accounts 1 and 2
	s.FundAccount(address1, sdk.NewCoin(stakeDenom, sdkmath.NewInt(v14.Account1VestingUstrd)))
	s.FundAccount(address2, sdk.NewCoin(stakeDenom, sdkmath.NewInt(v14.Account2VestingUstrd)))

	// Store the accounts as ContinuousVestingAccounts
	s.App.AccountKeeper.SetAccount(s.Ctx, account1)
	s.App.AccountKeeper.SetAccount(s.Ctx, account2)
}

func (s *UpgradeTestSuite) CheckVestingStoreAfterUpgrade() {
	afterCtx := s.Ctx.WithBlockHeight(dummyUpgradeHeight)
	address1, err := sdk.AccAddressFromBech32(v14.Account1)
	s.Require().NoError(err)
	address2, err := sdk.AccAddressFromBech32(v14.Account2)
	s.Require().NoError(err)
	// Verify account1 is now a ClawbackVestingAccount
	account1 := s.App.AccountKeeper.GetAccount(afterCtx, address1)
	s.Require().IsType(&evmosvestingtypes.ClawbackVestingAccount{}, account1)
	// Verify account2 is still a ContinuousVestingAccount
	account2 := s.App.AccountKeeper.GetAccount(afterCtx, address2)
	s.Require().IsType(&types.ContinuousVestingAccount{}, account2)
	// Verify the vested tokens are correct
	// TODO
}

// ---------------------- Utils ----------------------
func initBaseAccount(address sdk.AccAddress, coins int64) (*authtypes.BaseAccount, sdk.Coins) {
	origCoins := sdk.Coins{sdk.NewInt64Coin(stakeDenom, coins)}
	bacc := authtypes.NewBaseAccountWithAddress(address)
	return bacc, origCoins
}

func (s *UpgradeTestSuite) CreateContinuousVestingAccount(address sdk.AccAddress, start int64, end int64, coins int64) *types.ContinuousVestingAccount {
	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)

	// init a base account
	// send tokens to the base account
	bacc, origCoins := initBaseAccount(address, coins)
	cva := types.NewContinuousVestingAccount(bacc, origCoins, start, end)

	// Sanity check the vesting schedule
	// require no coins vested in the very beginning of the vesting schedule
	vestedCoins := cva.GetVestedCoins(startTime)
	s.Require().Nil(vestedCoins)

	// require all coins vested at the end of the vesting schedule)
	vestedCoins = cva.GetVestedCoins(endTime)
	s.Require().Equal(origCoins, vestedCoins)

	// require 50% of coins vested
	midpoint := time.Unix((start+end)/2, 0)
	vestedCoins = cva.GetVestedCoins(midpoint)
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(stakeDenom, coins/2)}, vestedCoins)

	return cva
}
