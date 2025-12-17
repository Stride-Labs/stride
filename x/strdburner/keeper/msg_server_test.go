package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	claimvestingtypes "github.com/Stride-Labs/stride/v31/x/claim/vesting/types"

	"github.com/Stride-Labs/stride/v31/x/strdburner/keeper"
	"github.com/Stride-Labs/stride/v31/x/strdburner/types"
)

func (s *KeeperTestSuite) verifyUserBurnEvents(address sdk.AccAddress, amount sdkmath.Int) {
	s.CheckEventValueEmitted(types.EventTypeUserBurn, types.AttributeAddress, address.String())
	s.CheckEventValueEmitted(types.EventTypeUserBurn, types.AttributeAmount, sdk.NewCoin(keeper.USTRD, amount).String())
}

// --------------------------------------------
//            Generic Burn Tests
// --------------------------------------------

// Test successful burns across multiple accounts
func (s *KeeperTestSuite) TestBurns_Successful_MultipleUsers() {
	initialBalance := sdkmath.NewInt(10_000)

	// Fund each account
	acc1, acc2, acc3 := s.TestAccs[0], s.TestAccs[1], s.TestAccs[2]
	s.FundAccount(acc1, sdk.NewCoin(keeper.USTRD, initialBalance))
	s.FundAccount(acc2, sdk.NewCoin(keeper.USTRD, initialBalance))
	s.FundAccount(acc3, sdk.NewCoin(keeper.USTRD, initialBalance))

	// Build valid messages
	msg1 := types.MsgBurn{
		Burner: acc1.String(),
		Amount: sdkmath.NewInt(1000),
	}
	msg2 := types.MsgBurn{
		Burner: acc2.String(),
		Amount: sdkmath.NewInt(2000),
	}
	msg3 := types.MsgBurn{
		Burner: acc3.String(),
		Amount: sdkmath.NewInt(3000),
	}

	// Execute each burn message and the respective events
	_, err := s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg1)
	s.Require().NoError(err)
	s.verifyUserBurnEvents(acc1, msg1.Amount)

	_, err = s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg2)
	s.Require().NoError(err)
	s.verifyUserBurnEvents(acc2, msg2.Amount)

	_, err = s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg3)
	s.Require().NoError(err)
	s.verifyUserBurnEvents(acc3, msg3.Amount)

	// Verify cumulative accounting
	expectedTotalBurned := msg1.Amount.Add(msg2.Amount).Add(msg3.Amount)

	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	s.Require().Equal(expectedTotalBurned, totalBurned)

	totalUserBurned := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	s.Require().Equal(expectedTotalBurned, totalUserBurned)

	protocolBurned := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	s.Require().Equal(sdkmath.ZeroInt(), protocolBurned)

	// Verify user accounting
	userBurned1 := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1)
	userBurned2 := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2)
	userBurned3 := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc3)

	s.Require().Equal(msg1.Amount, userBurned1)
	s.Require().Equal(msg2.Amount, userBurned2)
	s.Require().Equal(msg3.Amount, userBurned3)

	// Verify user balances
	userBalance1 := s.App.BankKeeper.GetBalance(s.Ctx, acc1, keeper.USTRD)
	userBalance2 := s.App.BankKeeper.GetBalance(s.Ctx, acc2, keeper.USTRD)
	userBalance3 := s.App.BankKeeper.GetBalance(s.Ctx, acc3, keeper.USTRD)

	s.Require().Equal(userBalance1.Amount, initialBalance.Sub(msg1.Amount))
	s.Require().Equal(userBalance2.Amount, initialBalance.Sub(msg2.Amount))
	s.Require().Equal(userBalance3.Amount, initialBalance.Sub(msg3.Amount))

	// Verify token supply
	expectedSupply := initialBalance.Mul(sdkmath.NewInt(3)).Sub(totalBurned)
	actualSupply := s.App.BankKeeper.GetSupply(s.Ctx, keeper.USTRD)
	s.Require().Equal(actualSupply.Amount, expectedSupply)
}

// Test a failed burn from an insufficient STRD balance
func (s *KeeperTestSuite) TestBurn_Successful_MultipleBurnsFromUser() {
	initialBalance := sdkmath.NewInt(10_000)

	acc := s.TestAccs[0]
	s.FundAccount(acc, sdk.NewCoin(keeper.USTRD, initialBalance))

	// Build valid burn messages
	msg1 := types.MsgBurn{
		Burner: acc.String(),
		Amount: sdkmath.NewInt(1000),
	}
	msg2 := types.MsgBurn{
		Burner: acc.String(),
		Amount: sdkmath.NewInt(2000),
	}
	msg3 := types.MsgBurn{
		Burner: acc.String(),
		Amount: sdkmath.NewInt(3000),
	}

	// Burn 3 times from the same user
	_, err := s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg1)
	s.Require().NoError(err)
	s.verifyUserBurnEvents(acc, msg1.Amount)

	_, err = s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg2)
	s.Require().NoError(err)
	s.verifyUserBurnEvents(acc, msg2.Amount)

	_, err = s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg3)
	s.Require().NoError(err)
	s.verifyUserBurnEvents(acc, msg3.Amount)

	// Check global state
	expectedTotalBurned := msg1.Amount.Add(msg2.Amount).Add(msg3.Amount)

	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	s.Require().Equal(expectedTotalBurned, totalBurned)

	totalUserBurned := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	s.Require().Equal(expectedTotalBurned, totalUserBurned)

	protocolBurned := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	s.Require().Equal(sdkmath.ZeroInt(), protocolBurned)

	// Check user state
	userBalance := s.App.BankKeeper.GetBalance(s.Ctx, acc, keeper.USTRD)
	s.Require().Equal(userBalance.Amount, initialBalance.Sub(expectedTotalBurned))
}

// Test a failed burn from an insufficient STRD balance
func (s *KeeperTestSuite) TestBurn_Failed_InsufficientBalance() {
	acc := s.TestAccs[0]
	s.FundAccount(acc, sdk.NewCoin(keeper.USTRD, sdkmath.NewInt(999)))

	msg := types.MsgBurn{
		Burner: acc.String(),
		Amount: sdkmath.NewInt(1000),
	}

	_, err := s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// --------------------------------------------
//            Vesting Test Helpers
// --------------------------------------------

type VestingState int

const (
	VestingStateParitallyVested VestingState = iota
	VestingStateFullyVested
)

// Helper to create a new base account
func (s *KeeperTestSuite) CreateNewBaseAccount(account sdk.AccAddress) *authtypes.BaseAccount {
	nextAccountNumber, err := s.App.AccountKeeper.AccountNumber.Next(s.Ctx)
	s.Require().NoError(err)

	baseAccount := authtypes.NewBaseAccountWithAddress(account)
	baseAccount.AccountNumber = nextAccountNumber + 1

	return baseAccount
}

// Helper to get start and end times of a vesting account
func (s *KeeperTestSuite) GetVestingTimes(vestingState VestingState) (int64, int64) {
	currentTime := s.Ctx.BlockTime().Unix()
	if vestingState == VestingStateParitallyVested {
		return currentTime - 10_000, currentTime + 10_000
	}
	if vestingState == VestingStateFullyVested {
		return currentTime - 10_000, currentTime
	}
	panic("Unrecognized vesting state")
}

// Helper to confirm tokens are locked in a continuous vesting account
func (s *KeeperTestSuite) ConfirmContinuousVestingLockedTokens(address sdk.AccAddress, expectedBalance sdkmath.Int) {
	vestingAccount := s.App.AccountKeeper.GetAccount(s.Ctx, address).(*vestingtypes.ContinuousVestingAccount)
	vestedBalance := vestingAccount.LockedCoins(s.Ctx.BlockTime()).AmountOf(keeper.USTRD)
	s.Require().Equal(expectedBalance, vestedBalance, "locked balance")
}

// Helper to confirm tokens are locked in a periodic vesting account
func (s *KeeperTestSuite) ConfirmPeriodicVestingLockedTokens(address sdk.AccAddress, expectedBalance sdkmath.Int) {
	vestingAccount := s.App.AccountKeeper.GetAccount(s.Ctx, address).(*vestingtypes.PeriodicVestingAccount)
	vestedBalance := vestingAccount.LockedCoins(s.Ctx.BlockTime()).AmountOf(keeper.USTRD)
	s.Require().Equal(expectedBalance, vestedBalance, "locked balance")
}

// Helper to confirm tokens are locked in a periodic vesting account
func (s *KeeperTestSuite) ConfirmStridePeriodicVestingLockedTokens(address sdk.AccAddress, expectedBalance sdkmath.Int) {
	vestingAccount := s.App.AccountKeeper.GetAccount(s.Ctx, address).(*claimvestingtypes.StridePeriodicVestingAccount)
	vestedBalance := vestingAccount.LockedCoins(s.Ctx.BlockTime()).AmountOf(keeper.USTRD)
	s.Require().Equal(expectedBalance, vestedBalance, "locked balance")
}

// Helper to create a new continuous vesting account that's partially vested
func (s *KeeperTestSuite) CreateContinuousVestingAccount(
	address sdk.AccAddress,
	initialBalance sdkmath.Int,
	vestingState VestingState,
) {
	baseAccount := s.CreateNewBaseAccount(address)
	s.FundAccount(address, sdk.NewCoin(keeper.USTRD, initialBalance))

	initialCoins := sdk.NewCoins(sdk.NewCoin(keeper.USTRD, initialBalance))
	startTime, endTime := s.GetVestingTimes(vestingState)

	vestingAccount, err := vestingtypes.NewContinuousVestingAccount(baseAccount, initialCoins, startTime, endTime)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, vestingAccount)

	if vestingState == VestingStateParitallyVested {
		s.ConfirmContinuousVestingLockedTokens(address, initialBalance.Quo(sdkmath.NewInt(2))) // should be 50% vested
	} else if vestingState == VestingStateFullyVested {
		s.ConfirmContinuousVestingLockedTokens(address, sdkmath.ZeroInt())
	}
}

// Helper to create a new continuous vesting account that's partially vested
func (s *KeeperTestSuite) CreatePeriodicVestingAccount(
	address sdk.AccAddress,
	initialBalance sdkmath.Int,
	vestingState VestingState,
) {
	baseAccount := s.CreateNewBaseAccount(address)
	s.FundAccount(address, sdk.NewCoin(keeper.USTRD, initialBalance))

	initialCoins := sdk.NewCoins(sdk.NewCoin(keeper.USTRD, initialBalance))
	startTime, endTime := s.GetVestingTimes(vestingState)

	halfCoins := initialCoins.QuoInt(sdkmath.NewInt(2))
	periods := vestingtypes.Periods{
		vestingtypes.Period{Length: 1, Amount: halfCoins},
		vestingtypes.Period{Length: endTime - startTime - 1, Amount: halfCoins},
	}

	vestingAccount, err := vestingtypes.NewPeriodicVestingAccount(baseAccount, initialCoins, startTime, periods)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, vestingAccount)

	if vestingState == VestingStateParitallyVested {
		s.ConfirmPeriodicVestingLockedTokens(address, initialBalance.Quo(sdkmath.NewInt(2))) // should be 50% vested
	} else if vestingState == VestingStateFullyVested {
		s.ConfirmPeriodicVestingLockedTokens(address, sdkmath.ZeroInt())
	}
}

// Helper to create a new continuous vesting account that's partially vested
func (s *KeeperTestSuite) CreateStridePeriodicVestingAccount(
	address sdk.AccAddress,
	initialBalance sdkmath.Int,
	vestingState VestingState,
) {
	baseAccount := s.CreateNewBaseAccount(address)
	s.FundAccount(address, sdk.NewCoin(keeper.USTRD, initialBalance))

	initialCoins := sdk.NewCoins(sdk.NewCoin(keeper.USTRD, initialBalance))
	startTime, endTime := s.GetVestingTimes(vestingState)

	periods := claimvestingtypes.Periods{
		claimvestingtypes.Period{StartTime: startTime, Length: endTime - startTime, Amount: initialCoins},
	}

	vestingAccount := claimvestingtypes.NewStridePeriodicVestingAccount(baseAccount, initialCoins, periods)
	s.App.AccountKeeper.SetAccount(s.Ctx, vestingAccount)

	if vestingState == VestingStateParitallyVested {
		s.ConfirmStridePeriodicVestingLockedTokens(address, initialBalance.Quo(sdkmath.NewInt(2))) // should be 50% vested
	} else if vestingState == VestingStateFullyVested {
		s.ConfirmStridePeriodicVestingLockedTokens(address, sdkmath.ZeroInt())
	}
}

// Helper to submit a burn tx
func (s *KeeperTestSuite) TryBurn(address sdk.AccAddress, amount sdkmath.Int) error {
	msg := types.MsgBurn{
		Burner: address.String(),
		Amount: amount,
	}
	_, err := s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg)
	return err
}

// --------------------------------------------
//     Continous Vesting Account Burn Tests
// --------------------------------------------

// Tests burning from a continuous vesting account that has insufficient unlocked balance,
// causing the account to get downgraded in order to fulfill the burn
func (s *KeeperTestSuite) TestBurn_ContinuousVestingAccount_PartiallyVested_Downgraded() {
	// The account will have 1000 tokens, with 500 vested
	// We'll try to burn 900 which will force a downgrade for the unlocked tokens
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(900)

	// Create continuous vesting account
	s.CreateContinuousVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Attempt to burn into the locked portion
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it's no longer a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*vestingtypes.ContinuousVestingAccount)
	s.Require().False(ok, "Should no longer be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().True(ok, "Should be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests burning from a continuous vesting account that has enough unlocked tokens
// to cover the burn, and thus does not need to be downgraded
func (s *KeeperTestSuite) TestBurn_ContinuousVestingAccount_PartiallyVested_NotDowngraded() {
	// The account will have 1000 tokens, with 500 vested
	// We'll try to burn 100 which will force a downgrade for the unlocked tokens
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(100)

	// Create continuous vesting account
	s.CreateContinuousVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it is still a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*vestingtypes.ContinuousVestingAccount)
	s.Require().True(ok, "Should still be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().False(ok, "Should not be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests buring from a continuous vesting account
func (s *KeeperTestSuite) TestBurn_ContinuousVestingAccount_PartiallyVested_InsufficientBalance() {
	// Attempt to burn with an amount greater than the balance
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(1001)

	// Create continuous vesting account
	s.CreateContinuousVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// Tests burning from a continuous vesting account that has everything vested, so it is not downgraded
func (s *KeeperTestSuite) TestBurn_ContinuousVestingAccount_FullyVested_NotDowngraded() {
	// The account will have 1000 tokens, all vested
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(100)

	// Create continuous vesting account
	s.CreateContinuousVestingAccount(address, initialBalance, VestingStateFullyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it is still a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*vestingtypes.ContinuousVestingAccount)
	s.Require().True(ok, "Should still be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().False(ok, "Should not be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests buring from a continuous vesting account
func (s *KeeperTestSuite) TestBurn_ContinuousVestingAccount_FullyVested_InsufficientBalance() {
	// Attempt to burn with an amount greater than the balance
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(1001)

	// Create continuous vesting account
	s.CreateContinuousVestingAccount(address, initialBalance, VestingStateFullyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// --------------------------------------------
//	Periodic Vesting Account Burn Tests
// --------------------------------------------

// Tests burning from a periodic vesting account that has insufficient unlocked balance,
// causing the account to get downgraded in order to fulfill the burn
func (s *KeeperTestSuite) TestBurn_PeriodicVestingAccount_PartiallyVested_Downgraded() {
	// The account will have 1000 tokens, with 500 vested
	// We'll try to burn 900 which will force a downgrade for the unlocked tokens
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(900)

	// Create periodic vesting account
	s.CreatePeriodicVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Attempt to burn into the locked portion
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it's no longer a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*vestingtypes.PeriodicVestingAccount)
	s.Require().False(ok, "Should no longer be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().True(ok, "Should be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests burning from a periodic vesting account that has enough unlocked tokens
// to cover the burn, and thus does not need to be downgraded
func (s *KeeperTestSuite) TestBurn_PeriodicVestingAccount_PartiallyVested_NotDowngraded() {
	// The account will have 1000 tokens, with 500 vested
	// We'll try to burn 100 which will force a downgrade for the unlocked tokens
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(100)

	// Create periodic vesting account
	s.CreatePeriodicVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it is still a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*vestingtypes.PeriodicVestingAccount)
	s.Require().True(ok, "Should still be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().False(ok, "Should not be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests buring from a periodic vesting account
func (s *KeeperTestSuite) TestBurn_PeriodicVestingAccount_PartiallyVested_InsufficientBalance() {
	// Attempt to burn with an amount greater than the balance
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(1001)

	// Create periodic vesting account
	s.CreatePeriodicVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// Tests burning from a periodic vesting account that has everything vested, so it is not downgraded
func (s *KeeperTestSuite) TestBurn_PeriodicVestingAccount_FullyVested_NotDowngraded() {
	// The account will have 1000 tokens, all vested
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(100)

	// Create periodic vesting account
	s.CreatePeriodicVestingAccount(address, initialBalance, VestingStateFullyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it is still a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*vestingtypes.PeriodicVestingAccount)
	s.Require().True(ok, "Should still be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().False(ok, "Should not be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests buring from a periodic vesting account
func (s *KeeperTestSuite) TestBurn_PeriodicVestingAccount_FullyVested_InsufficientBalance() {
	// Attempt to burn with an amount greater than the balance
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(1001)

	// Create periodic vesting account
	s.CreatePeriodicVestingAccount(address, initialBalance, VestingStateFullyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// --------------------------------------------
//	Stride Periodic Vesting Account Burn Tests
// --------------------------------------------

// Tests burning from a periodic vesting account that has insufficient unlocked balance,
// causing the account to get downgraded in order to fulfill the burn
func (s *KeeperTestSuite) TestBurn_StridePeriodicVestingAccount_PartiallyVested_Downgraded() {
	// The account will have 1000 tokens, with 500 vested
	// We'll try to burn 900 which will force a downgrade for the unlocked tokens
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(900)

	// Create periodic vesting account
	s.CreateStridePeriodicVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Attempt to burn into the locked portion
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it's no longer a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*claimvestingtypes.StridePeriodicVestingAccount)
	s.Require().False(ok, "Should no longer be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().True(ok, "Should be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests burning from a periodic vesting account that has enough unlocked tokens
// to cover the burn, and thus does not need to be downgraded
func (s *KeeperTestSuite) TestBurn_StridePeriodicVestingAccount_PartiallyVested_NotDowngraded() {
	// The account will have 1000 tokens, with 500 vested
	// We'll try to burn 100 which will force a downgrade for the unlocked tokens
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(100)

	// Create periodic vesting account
	s.CreateStridePeriodicVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it is still a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*claimvestingtypes.StridePeriodicVestingAccount)
	s.Require().True(ok, "Should still be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().False(ok, "Should not be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests buring from a periodic vesting account
func (s *KeeperTestSuite) TestBurn_StridePeriodicVestingAccount_PartiallyVested_InsufficientBalance() {
	// Attempt to burn with an amount greater than the balance
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(1001)

	// Create periodic vesting account
	s.CreateStridePeriodicVestingAccount(address, initialBalance, VestingStateParitallyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// Tests burning from a periodic vesting account that has everything vested, so it is not downgraded
func (s *KeeperTestSuite) TestBurn_StridePeriodicVestingAccount_FullyVested_NotDowngraded() {
	// The account will have 1000 tokens, all vested
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(100)

	// Create periodic vesting account
	s.CreateStridePeriodicVestingAccount(address, initialBalance, VestingStateFullyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address)
	s.Require().Equal(burnAmount, actualBurnAmount, "Burn amount by address")

	// Confirm it is still a vesting account
	accountI := s.App.AccountKeeper.GetAccount(s.Ctx, address)
	_, ok := accountI.(*claimvestingtypes.StridePeriodicVestingAccount)
	s.Require().True(ok, "Should still be a vesting account")

	_, ok = accountI.(*authtypes.BaseAccount)
	s.Require().False(ok, "Should not be a base account")

	// Confirm updated balance
	updatedBalance := s.App.BankKeeper.GetBalance(s.Ctx, address, keeper.USTRD).Amount
	s.Require().Equal(initialBalance.Sub(burnAmount), updatedBalance, "updated balance")
}

// Tests buring from a periodic vesting account
func (s *KeeperTestSuite) TestBurn_StridePeriodicVestingAccount_FullyVested_InsufficientBalance() {
	// Attempt to burn with an amount greater than the balance
	address := s.TestAccs[0]
	initialBalance := sdkmath.NewInt(1000)
	burnAmount := sdkmath.NewInt(1001)

	// Create periodic vesting account
	s.CreateStridePeriodicVestingAccount(address, initialBalance, VestingStateFullyVested)

	// Burn some of the locked tokens
	err := s.TryBurn(address, burnAmount)
	s.Require().ErrorContains(err, "unable to transfer tokens to the burner module account")
}

// --------------------------------------------
//	              Link Tests
// --------------------------------------------

func (s *KeeperTestSuite) TestLink_Successful() {
	strideAddress := s.TestAccs[0]
	linkedAddress1 := "0x1"
	linkedAddress2 := "0x1"

	// Link one address
	msg := types.MsgLink{
		StrideAddress: strideAddress.String(),
		LinkedAddress: linkedAddress1,
	}
	_, err := s.GetMsgServer().Link(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when linking")
	s.Require().Equal(linkedAddress1, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, strideAddress))

	// Link a different address (overwrites the first)
	msg.LinkedAddress = linkedAddress2
	_, err = s.GetMsgServer().Link(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when linking")
	s.Require().Equal(linkedAddress2, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, strideAddress))
}
