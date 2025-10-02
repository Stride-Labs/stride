package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/strdburner/keeper"
	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

func (s *KeeperTestSuite) verifyUserBurnEvents(address sdk.AccAddress, amount sdkmath.Int) {
	s.CheckEventValueEmitted(types.EventTypeUserBurn, types.AttributeAddress, address.String())
	s.CheckEventValueEmitted(types.EventTypeUserBurn, types.AttributeAmount, sdk.NewCoin(keeper.USTRD, amount).String())
}

// Test successful burns across multiple accounts
func (s *KeeperTestSuite) TestSuccessfulBurns() {
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
func (s *KeeperTestSuite) TestBurnFailed_InsufficientBalance() {

}
