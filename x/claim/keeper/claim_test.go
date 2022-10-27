package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

// Test functionality for loading allocation data(csv)
func (suite *KeeperTestSuite) TestLoadAllocationData() {
	suite.SetupTest()
	var allocations = `identifier,address,weight
stride,stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk,0.5
stride,stride1h4astdfzjhcwahtfrh24qtvndzzh49xvqtfftk,0.3`

	ok := suite.app.ClaimKeeper.LoadAllocationData(suite.ctx, allocations)
	suite.Require().True(ok)

	totalWeight, err := suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, "stride")
	suite.Require().NoError(err)
	suite.Require().True(totalWeight.Equal(sdk.MustNewDecFromStr("0.8")))

	addr, _ := sdk.AccAddressFromBech32("stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk")
	claimRecord, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr, "stride")
	suite.Require().Equal(claimRecord.Address, "stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk")
	suite.Require().True(claimRecord.Weight.Equal(sdk.MustNewDecFromStr("0.5")))
	suite.Require().Equal(claimRecord.ActionCompleted, []bool{false, false, false})
}

// Check unclaimable account's balance after staking
func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupTest()

	// Set a normal user account
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	claim, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1, "stride")
	suite.NoError(err)
	suite.Equal(types.ClaimRecord{}, claim)

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)

	// Get balances for the account
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(sdk.Coins{}, balances)
}

//Check balances before and after airdrop starts
func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupTest()

	airdropStartTime := time.Now().Add(time.Hour)

	err := suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
		Airdrops: []*types.Airdrop{
			{
				AirdropIdentifier:  types.DefaultAirdropIdentifier,
				AirdropStartTime:   airdropStartTime,
				AirdropDuration:    types.DefaultAirdropDuration,
				ClaimDenom:         sdk.DefaultBondDenom,
				DistributorAddress: distributors[types.DefaultAirdropIdentifier].String(),
			},
		},
	})
	suite.Require().NoError(err)

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	val1 := sdk.ValAddress(addr1)

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}
	suite.app.ClaimKeeper.SetTotalWeight(suite.ctx, claimRecords[0].Weight, "stride")

	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	err = suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride")
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	coins, err = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionFree, "stride")
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	suite.app.ClaimKeeper.AfterDelegationModified(suite.ctx, addr1, val1)
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is before starting air drop, so claim module should not send the balances to the user after swap.
	suite.True(balances.Empty())

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx.WithBlockTime(airdropStartTime), addr1)
	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user after liquid stake.
	claimableAmountForLiquidStake := sdk.NewDecWithPrec(60, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // 60% for liquid stake
	suite.Require().Equal(balances.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())
}

// Run all airdrop flow
func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}

	suite.app.ClaimKeeper.SetTotalWeight(suite.ctx, claimRecords[0].Weight, "stride")
	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride")
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)).String())

	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2, "stride")
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.Coins{})

	// get rewards amount for free
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionFree, "stride")
	suite.Require().NoError(err)
	claimableAmountForFree := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionDelegateStake, "stride")
	suite.Require().NoError(err)
	claimableAmountForStake := sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(20, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionLiquidStake, "stride")
	suite.Require().NoError(err)
	claimableAmountForLiquidStake := sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(80, 2)).
		Mul(sdk.NewDecWithPrec(60, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)).String())

	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx, "stride")
	suite.Require().NoError(err)

	// get spendable balances 2 months later
	ctx := suite.ctx.WithBlockTime(time.Now().Add(types.DefaultVestingDurationForDelegateStake))
	coins = suite.app.BankKeeper.SpendableCoins(ctx, addr1)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake/2)).String())
}

// Run multi chain airdrop flow
func (suite *KeeperTestSuite) TestMultiChainAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}

	suite.app.ClaimKeeper.SetTotalWeight(suite.ctx, claimRecords[0].Weight, "stride")
	suite.app.ClaimKeeper.SetTotalWeight(suite.ctx, claimRecords[1].Weight, "juno")
	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride")
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)).String())

	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2, "juno")
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	// get rewards amount for free (stride user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionFree, "stride")
	suite.Require().NoError(err)
	claimableAmountForFree := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake (stride user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionDelegateStake, "stride")
	suite.Require().NoError(err)
	claimableAmountForStake := sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(20, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake (stride user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionLiquidStake, "stride")
	suite.Require().NoError(err)
	claimableAmountForLiquidStake := sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(80, 2)).
		Mul(sdk.NewDecWithPrec(60, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)).String())

	suite.app.ClaimKeeper.EndBlocker(suite.ctx.WithBlockTime(time.Now().Add(types.DefaultAirdropDuration)))
	weight, err := suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().NoError(err)
	suite.Require().Equal(weight, sdk.ZeroDec())

	records := suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(0, len(records))

	//*********************** End of Stride airdrop *************************

	// claim airdrops for juno users after ending stride airdrop
	// get rewards amount for free (juno user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ActionFree, "juno")
	suite.Require().NoError(err)
	claimableAmountForFree = sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake (juno user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ActionDelegateStake, "juno")
	suite.Require().NoError(err)
	claimableAmountForStake = sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(20, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake (juno user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ActionLiquidStake, "juno")
	suite.Require().NoError(err)
	claimableAmountForLiquidStake = sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(80, 2)).
		Mul(sdk.NewDecWithPrec(60, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr2)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)).String())

	// after 3 years, juno users should be still able to claim
	suite.app.ClaimKeeper.ClearClaimedStatus(suite.ctx, "juno")
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ActionFree, "juno")
	suite.Require().NoError(err)

	claimableAmountForFree = sdk.NewDecWithPrec(80, 2).
		Mul(sdk.NewDecWithPrec(80, 2)).
		Mul(sdk.NewDecWithPrec(40, 2)).
		Mul(sdk.NewDecWithPrec(20, 2)).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100)*(40/100), claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// after 3 years + 1 hour, juno users shouldn't be able to claim anymore
	suite.app.ClaimKeeper.EndBlocker(suite.ctx.WithBlockTime(time.Now().Add(time.Hour).Add(types.DefaultAirdropDuration)))
	suite.app.ClaimKeeper.ClearClaimedStatus(suite.ctx, "juno")
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour).Add(types.DefaultAirdropDuration)), addr2, types.ActionFree, "juno")
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	weight, err = suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().NoError(err)
	suite.Require().Equal(weight, sdk.ZeroDec())

	records = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(0, len(records))
	//*********************** End of Juno airdrop *************************
}
