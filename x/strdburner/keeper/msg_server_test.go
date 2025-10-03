package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	claimvestingtypes "github.com/Stride-Labs/stride/v28/x/claim/vesting/types"

	"github.com/Stride-Labs/stride/v28/x/strdburner/keeper"
	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

func (s *KeeperTestSuite) verifyUserBurnEvents(address sdk.AccAddress, amount sdkmath.Int) {
	s.CheckEventValueEmitted(types.EventTypeUserBurn, types.AttributeAddress, address.String())
	s.CheckEventValueEmitted(types.EventTypeUserBurn, types.AttributeAmount, sdk.NewCoin(keeper.USTRD, amount).String())
}

func (s *KeeperTestSuite) CreateNewBaseAccount(account sdk.AccAddress) *authtypes.BaseAccount {
	nextAccountNumber, err := s.App.AccountKeeper.AccountNumber.Next(s.Ctx)
	s.Require().NoError(err)

	baseAccount := authtypes.NewBaseAccountWithAddress(account)
	baseAccount.AccountNumber = nextAccountNumber + 1

	return baseAccount
}

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

// Tests buring from a continuous vesting account
func (s *KeeperTestSuite) TestBurn_Successful_ContinuousVestingAccount() {
	account := s.TestAccs[0]

	initialBalance := sdkmath.NewInt(1000)
	initialCoins := sdk.NewCoins(sdk.NewCoin(keeper.USTRD, initialBalance))
	startTime := time.Now().Unix()
	endTime := time.Date(2035, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	// Create continuous vesting account
	baseAccount := s.CreateNewBaseAccount(account)
	vestingAccount, err := vestingtypes.NewContinuousVestingAccount(baseAccount, initialCoins, startTime, endTime)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, vestingAccount)

	// Confirm tokens are locked
	vestingAccount = s.App.AccountKeeper.GetAccount(s.Ctx, account).(*vestingtypes.ContinuousVestingAccount)
	vestedBalance := vestingAccount.LockedCoins(s.Ctx.BlockTime()).AmountOf(keeper.USTRD)
	s.Require().Equal(initialBalance, vestedBalance, "locked balance")

	// Burn some of the locked tokens
	expectedBurnAmount := sdkmath.NewInt(100)
	msg := types.MsgBurn{
		Burner: account.String(),
		Amount: expectedBurnAmount,
	}
	_, err = s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, account)
	s.Require().Equal(expectedBurnAmount, actualBurnAmount, "Burn amount by address")
}

// Tests buring from a periodic vesting account
func (s *KeeperTestSuite) TestBurn_Successful_PeriodicVestingAccount() {
	account := s.TestAccs[0]

	initialBalance := sdkmath.NewInt(1000)
	initialCoins := sdk.NewCoins(sdk.NewCoin(keeper.USTRD, initialBalance))
	startTime := time.Now().Unix()
	endTime := time.Date(2035, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	periods := vestingtypes.Periods{
		vestingtypes.Period{Length: endTime - startTime, Amount: initialCoins},
	}

	// Create stride periodic vesting account
	baseAccount := s.CreateNewBaseAccount(account)
	vestingAccount, err := vestingtypes.NewPeriodicVestingAccount(baseAccount, initialCoins, startTime, periods)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, vestingAccount)

	// Confirm tokens are locked
	vestingAccount = s.App.AccountKeeper.GetAccount(s.Ctx, account).(*vestingtypes.PeriodicVestingAccount)
	vestedBalance := vestingAccount.LockedCoins(s.Ctx.BlockTime()).AmountOf(keeper.USTRD)
	s.Require().Equal(initialBalance, vestedBalance, "locked balance")

	// Burn some of the locked tokens
	expectedBurnAmount := sdkmath.NewInt(100)
	msg := types.MsgBurn{
		Burner: account.String(),
		Amount: expectedBurnAmount,
	}
	_, err = s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, account)
	s.Require().Equal(expectedBurnAmount, actualBurnAmount, "Burn amount by address")
}

// Tests buring from a stride periodic vesting account
func (s *KeeperTestSuite) TestBurn_Successful_StridePeriodicVestingAccount() {
	account := s.TestAccs[0]

	initialBalance := sdkmath.NewInt(1000)
	initialCoins := sdk.NewCoins(sdk.NewCoin(keeper.USTRD, initialBalance))
	startTime := time.Now().Unix()
	endTime := time.Date(2035, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	periods := claimvestingtypes.Periods{
		claimvestingtypes.Period{StartTime: time.Now().Unix(), Length: endTime - startTime, Amount: initialCoins},
	}

	// Create stride periodic vesting account
	baseAccount := s.CreateNewBaseAccount(account)
	vestingAccount := claimvestingtypes.NewStridePeriodicVestingAccount(baseAccount, initialCoins, periods)
	s.App.AccountKeeper.SetAccount(s.Ctx, vestingAccount)

	// Confirm tokens are locked
	vestingAccount = s.App.AccountKeeper.GetAccount(s.Ctx, account).(*claimvestingtypes.StridePeriodicVestingAccount)
	vestedBalance := vestingAccount.LockedCoins(s.Ctx.BlockTime()).AmountOf(keeper.USTRD)
	s.Require().Equal(initialBalance, vestedBalance, "locked balance")

	// Burn some of the locked tokens
	expectedBurnAmount := sdkmath.NewInt(100)
	msg := types.MsgBurn{
		Burner: account.String(),
		Amount: expectedBurnAmount,
	}
	_, err := s.GetMsgServer().Burn(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "burn should not have failed")

	// Confirm burn accounting change
	actualBurnAmount := s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, account)
	s.Require().Equal(expectedBurnAmount, actualBurnAmount, "Burn amount by address")
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
