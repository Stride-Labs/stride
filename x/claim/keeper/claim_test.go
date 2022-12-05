package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

// Test functionality for loading allocation data(csv)
func (suite *KeeperTestSuite) TestLoadAllocationData() {
	suite.SetupTest()
	var allocations = `identifier,address,weight
osmosis,osmo1g7yxhuppp5x3yqkah5mw29eqq5s4sv2fp6e2eg,0.5
osmosis,osmo1h4astdfzjhcwahtfrh24qtvndzzh49xvtm69fg,0.3
stride,stride1av5lwh0msnafn04xkhdyk6mrykxthrawy7uf3d,0.7
stride,stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk,0.3
stride,stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk,0.5`

	ok := suite.app.ClaimKeeper.LoadAllocationData(suite.ctx, allocations)
	suite.Require().True(ok)

	totalWeight, err := suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, "osmosis")
	suite.Require().NoError(err)
	suite.Require().True(totalWeight.Equal(sdk.MustNewDecFromStr("0.8")))

	totalWeight, err = suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, "stride")
	suite.Require().NoError(err)
	suite.Require().True(totalWeight.Equal(sdk.MustNewDecFromStr("1")))

	addr, _ := sdk.AccAddressFromBech32("stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk") // hex(stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk) = hex(osmo1g7yxhuppp5x3yqkah5mw29eqq5s4sv2fp6e2eg)
	claimRecord, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr, "osmosis")
	suite.Require().NoError(err)
	suite.Require().Equal(claimRecord.Address, "stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk")
	suite.Require().True(claimRecord.Weight.Equal(sdk.MustNewDecFromStr("0.5")))
	suite.Require().Equal(claimRecord.ActionCompleted, []bool{false, false, false})

	claimRecord, err = suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr, "stride")
	suite.Require().NoError(err)
	suite.Require().True(claimRecord.Weight.Equal(sdk.MustNewDecFromStr("0.3")))
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
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	err = suite.app.ClaimKeeper.SetClaimRecordsWithWeights(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride", false)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	coins, err = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ACTION_FREE, "stride", false)
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

// Check original user balances after being converted into stride vesting account
func (suite *KeeperTestSuite) TestBalancesAfterAccountConversion() {
	suite.SetupTest()

	// set a normal account
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

	initialBal := int64(1000)
	err := suite.app.BankKeeper.SendCoins(suite.ctx, distributors["stride"], addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
	suite.Require().NoError(err)

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}

	err = suite.app.ClaimKeeper.SetClaimRecordsWithWeights(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	// check if original account tokens are not affected after stride vesting
	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr, types.ACTION_DELEGATE_STAKE, "stride")
	suite.Require().NoError(err)
	claimableAmountForStake := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000 - initialBal)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake

	coinsBal := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
	suite.Require().Equal(coinsBal.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal+claimableAmountForStake)).String())

	spendableCoinsBal := suite.app.BankKeeper.SpendableCoins(suite.ctx, addr)
	suite.Require().Equal(spendableCoinsBal.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)).String())
}

// Run all airdrop flow
func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	weight := sdk.NewDecWithPrec(50, 2)

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            weight, // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            weight, // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}

	err := suite.app.ClaimKeeper.SetClaimRecordsWithWeights(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride", false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50_000_000)).String())

	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2, "stride", false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50_000_000)).String())

	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr3, "stride", false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.Coins{})

	// get rewards amount for free
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_FREE, "stride")
	suite.Require().NoError(err)
	claimableAmountForFree := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_DELEGATE_STAKE, "stride")
	suite.Require().NoError(err)
	claimableAmountForStake := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance is 90000000, claim 20% for stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_LIQUID_STAKE, "stride")
	suite.Require().NoError(err)
	claimableAmountForLiquidStake := sdk.NewDecWithPrec(60, 2).
		Mul(sdk.NewDec(100_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance = 80000000, claim 60% for liquid stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)).String())

	// get spendable balances 3 months later
	ctx := suite.ctx.WithBlockTime(time.Now().Add(types.DefaultVestingDurationForDelegateStake))
	coinsSpendable := suite.app.BankKeeper.SpendableCoins(ctx, addr1)
	suite.Require().Equal(coins.String(), coinsSpendable.String())

	// check if claims don't vest after initial period of 3 months
	suite.ctx = suite.ctx.WithBlockTime(time.Now().Add(types.DefaultVestingInitialPeriod))
	err = suite.app.ClaimKeeper.ResetClaimStatus(suite.ctx, "stride")
	suite.Require().NoError(err)
	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_LIQUID_STAKE, "stride")
	suite.Require().NoError(err)
	claimableAmountForLiquidStake2 := sdk.NewDecWithPrec(60, 2).
		Mul(sdk.NewDec(50_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance = 50000000*(60/100), claim 60% for liquid stake

	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake+claimableAmountForLiquidStake2)).String())
	coinsSpendable = suite.app.BankKeeper.SpendableCoins(ctx, addr1)
	suite.Require().Equal(coins.String(), coinsSpendable.String())

	// end airdrop
	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx, "stride")
	suite.Require().NoError(err)
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
		{
			Address:           addr1.String(),
			Weight:            sdk.NewDecWithPrec(30, 2), // 30%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "osmosis",
		},
	}

	err := suite.app.ClaimKeeper.SetClaimRecordsWithWeights(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride", false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)).String())

	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2, "juno", false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	identifiers := suite.app.ClaimKeeper.GetAirdropIdentifiersForUser(suite.ctx, addr1)
	suite.Require().Equal(identifiers[0], types.DefaultAirdropIdentifier)
	suite.Require().Equal(identifiers[1], "osmosis")

	// get rewards amount for free (stride, osmosis addresses)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_FREE, "stride")
	suite.Require().NoError(err)

	coins1, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_FREE, "osmosis")
	suite.Require().NoError(err)

	claimableAmountForFree := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake (stride, osmosis addresses)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_DELEGATE_STAKE, "stride")
	suite.Require().NoError(err)

	coins1, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_DELEGATE_STAKE, "osmosis")
	suite.Require().NoError(err)

	claimableAmountForStake := sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake (stride, osmosis addresses)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_LIQUID_STAKE, "stride")
	suite.Require().NoError(err)

	coins1, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ACTION_LIQUID_STAKE, "osmosis")
	suite.Require().NoError(err)

	claimableAmountForLiquidStake := sdk.NewDecWithPrec(60, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, (claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)*2)).String())

	// Verify that the max claimable amount is unchanged, even after claims
	maxCoins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride", true)
	suite.Require().NoError(err)
	suite.Require().Equal(maxCoins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)).String())
	claimableCoins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1, "stride", false)
	suite.Require().NoError(err)
	suite.Require().Equal(claimableCoins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	// check if stride and osmosis airdrops ended properly
	suite.ctx = suite.ctx.WithBlockHeight(1000)
	suite.app.ClaimKeeper.EndBlocker(suite.ctx.WithBlockTime(time.Now().Add(types.DefaultAirdropDuration)))
	// for stride
	weight, err := suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().NoError(err)
	suite.Require().Equal(weight, sdk.ZeroDec())

	records := suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(0, len(records))

	// for osmosis
	weight, err = suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, "osmosis")
	suite.Require().NoError(err)
	suite.Require().Equal(weight, sdk.ZeroDec())

	records = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, "osmosis")
	suite.Require().Equal(0, len(records))

	//*********************** End of Stride, Osmosis airdrop *************************

	// claim airdrops for juno users after ending stride airdrop
	// get rewards amount for stake (juno user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ACTION_DELEGATE_STAKE, "juno")
	suite.Require().NoError(err)
	claimableAmountForStake = sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake (juno user)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ACTION_LIQUID_STAKE, "juno")
	suite.Require().NoError(err)
	claimableAmountForLiquidStake = sdk.NewDecWithPrec(60, 2).
		Mul(sdk.NewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr2)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake+claimableAmountForLiquidStake)).String())

	// after 3 years, juno users should be still able to claim
	suite.ctx = suite.ctx.WithBlockTime(time.Now().Add(types.DefaultAirdropDuration))
	err = suite.app.ClaimKeeper.ResetClaimStatus(suite.ctx, "juno")
	suite.Require().NoError(err)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr2, types.ACTION_FREE, "juno")
	suite.Require().NoError(err)

	claimableAmountForFree = sdk.NewDecWithPrec(20, 2).
		Mul(sdk.NewDec(20_000_000)).
		RoundInt64() // remaining balance = 20000000*(20/100), claim 20% for free
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// after 3 years + 1 hour, juno users shouldn't be able to claim anymore
	suite.ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Hour).Add(types.DefaultAirdropDuration))
	suite.app.ClaimKeeper.EndBlocker(suite.ctx)
	err = suite.app.ClaimKeeper.ResetClaimStatus(suite.ctx, "juno")
	suite.Require().NoError(err)
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(time.Now().Add(time.Hour).Add(types.DefaultAirdropDuration)), addr2, types.ACTION_FREE, "juno")
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	weight, err = suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().NoError(err)
	suite.Require().Equal(weight, sdk.ZeroDec())

	records = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(0, len(records))
	//*********************** End of Juno airdrop *************************
}
