package keeper_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/utils"
	claimkeeper "github.com/Stride-Labs/stride/v9/x/claim/keeper"

	"github.com/Stride-Labs/stride/v9/x/claim/types"
	stridevestingtypes "github.com/Stride-Labs/stride/v9/x/claim/vesting/types"
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

// Check balances before and after airdrop starts
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

	err = suite.app.ClaimKeeper.AfterDelegationModified(suite.ctx, addr1, val1)
	suite.NoError(err)
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

// Check original user balances after being converted into stride vesting account
func (suite *KeeperTestSuite) TestClaimAccountTypes() {
	suite.SetupTest()

	// set a normal account
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	// Base Account can claim
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

	// Verify the account type has changed to stride vesting account
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr)
	_, isVestingAcc := acc.(*stridevestingtypes.StridePeriodicVestingAccount)
	suite.Require().True(isVestingAcc)

	// Initialize vesting accounts
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	account := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr2)
	err = suite.app.BankKeeper.SendCoins(suite.ctx, distributors["stride"], addr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
	suite.Require().NoError(err)
	suite.app.AccountKeeper.SetAccount(suite.ctx, vestingtypes.NewBaseVestingAccount(account.(*authtypes.BaseAccount), nil, 0))

	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	account = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr3)
	err = suite.app.BankKeeper.SendCoins(suite.ctx, distributors["stride"], addr3, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
	suite.Require().NoError(err)
	suite.app.AccountKeeper.SetAccount(suite.ctx, vestingtypes.NewContinuousVestingAccount(account.(*authtypes.BaseAccount), nil, 0, 0))

	addr4 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	account = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr4)
	err = suite.app.BankKeeper.SendCoins(suite.ctx, distributors["stride"], addr4, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
	suite.Require().NoError(err)
	suite.app.AccountKeeper.SetAccount(suite.ctx, vestingtypes.NewPeriodicVestingAccount(account.(*authtypes.BaseAccount), nil, 0, nil))

	// Init claim records
	for _, addr := range []sdk.AccAddress{addr2, addr3, addr4} {
		claimRecords := []types.ClaimRecord{
			{
				Address:           addr.String(),
				Weight:            sdk.NewDecWithPrec(50, 2),
				ActionCompleted:   []bool{false, false, false},
				AirdropIdentifier: types.DefaultAirdropIdentifier,
			},
		}
		err = suite.app.ClaimKeeper.SetClaimRecordsWithWeights(suite.ctx, claimRecords)
		suite.Require().NoError(err)
	}

	// Try to claim tokens with each account type
	for _, addr := range []sdk.AccAddress{addr2, addr3, addr4} {
		_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr, types.ACTION_DELEGATE_STAKE, "stride")
		suite.Require().ErrorContains(err, "only BaseAccount and StridePeriodicVestingAccount can claim")
	}
}

// Run all airdrop flow
func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	weight := sdk.NewDecWithPrec(50, 2)

	for _, addr := range []sdk.AccAddress{addr1, addr2, addr3} {
		initialBal := int64(1000)
		err := suite.app.BankKeeper.SendCoins(suite.ctx, distributors["stride"], addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		suite.Require().NoError(err)
		err = suite.app.BankKeeper.SendCoins(suite.ctx, addr, distributors["stride"], sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		suite.Require().NoError(err)
	}

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

	for _, addr := range []sdk.AccAddress{addr1, addr2} {
		initialBal := int64(1000)
		err := suite.app.BankKeeper.SendCoins(suite.ctx, distributors["stride"], addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		suite.Require().NoError(err)
		err = suite.app.BankKeeper.SendCoins(suite.ctx, addr, distributors["stride"], sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		suite.Require().NoError(err)
	}

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

func (suite *KeeperTestSuite) TestAreAllTrue() {
	suite.Require().True(claimkeeper.AreAllTrue([]bool{true, true, true}))
	suite.Require().False(claimkeeper.AreAllTrue([]bool{true, false, true}))
	suite.Require().False(claimkeeper.AreAllTrue([]bool{false, false, false}))
}

func (suite *KeeperTestSuite) TestCurrentAirdropRound() {
	startTime := time.Now().Add(-50 * 24 * time.Hour) // 50 days ago
	round := claimkeeper.CurrentAirdropRound(startTime)
	suite.Require().Equal(1, round)

	startTime = time.Now().Add(-100 * 24 * time.Hour) // 100 days ago
	round = claimkeeper.CurrentAirdropRound(startTime)
	suite.Require().Equal(2, round)

	startTime = time.Now().Add(-130 * 24 * time.Hour) // 130 days ago
	round = claimkeeper.CurrentAirdropRound(startTime)
	suite.Require().Equal(3, round)
}

func (suite *KeeperTestSuite) TestGetClaimStatus() {
	addresses := apptesting.CreateRandomAccounts(2)
	address := addresses[0].String()
	otherAddress := addresses[1].String()

	// Add 5 airdrops
	airdrops := types.Params{
		Airdrops: []*types.Airdrop{
			{AirdropIdentifier: types.DefaultAirdropIdentifier},
			{AirdropIdentifier: "juno"},
			{AirdropIdentifier: "osmosis"},
			{AirdropIdentifier: "terra"},
			{AirdropIdentifier: "stargaze"},
		},
	}
	err := suite.app.ClaimKeeper.SetParams(suite.ctx, airdrops)
	suite.Require().NoError(err)

	// For the given user, add 4 claim records
	// Stride and Juno are incomplete
	// Osmosis and terra are complete
	// User is not eligible for stargaze
	claimRecords := []types.ClaimRecord{
		{
			Address:           address,
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           address,
			ActionCompleted:   []bool{false, true, false},
			AirdropIdentifier: "juno",
		},
		{
			Address:           address,
			ActionCompleted:   []bool{true, true, true},
			AirdropIdentifier: "osmosis",
		},
		{
			Address:           otherAddress, // different address
			ActionCompleted:   []bool{true, true, true},
			AirdropIdentifier: "terra",
		},
	}
	for _, claimRecord := range claimRecords {
		err := suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, claimRecord)
		suite.Require().NoError(err)
	}

	expectedClaimStatus := []types.ClaimStatus{
		{AirdropIdentifier: types.DefaultAirdropIdentifier, Claimed: false},
		{AirdropIdentifier: "juno", Claimed: false},
		{AirdropIdentifier: "osmosis", Claimed: true},
	}

	// Confirm status lines up with expectations
	status, err := suite.app.ClaimKeeper.GetClaimStatus(suite.ctx, sdk.MustAccAddressFromBech32(address))
	suite.Require().NoError(err, "no error expected when getting claim status")

	suite.Require().Equal(len(expectedClaimStatus), len(status), "number of airdrops")
	for i := 0; i < len(expectedClaimStatus); i++ {
		suite.Require().Equal(expectedClaimStatus[i].AirdropIdentifier, status[i].AirdropIdentifier, "airdrop ID for %d", i)
		suite.Require().Equal(expectedClaimStatus[i].AirdropIdentifier, status[i].AirdropIdentifier, "airdrop claimed for %i", i)
	}

}

func (suite *KeeperTestSuite) TestGetClaimMetadata() {
	// Update airdrop epochs timing
	now := time.Now().UTC()
	epochs := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
	for _, epoch := range epochs {
		// Each airdrop's round will end 10 days from now
		// Round 1 epoch started 80 days ago
		// Round 2 and 3 epochs have their current epoch started 20 days ago
		//  but had their genesis 110 and 140 days ago respectively
		epoch.CurrentEpochStartTime = now.Add(-20 * 24 * time.Hour) // 20 day ago

		switch epoch.Identifier {
		case "airdrop-" + types.DefaultAirdropIdentifier:
			epoch.Duration = time.Hour * 24 * 90                        // 90 days
			epoch.StartTime = now.Add(-80 * 24 * time.Hour)             // 80 days ago - round 1
			epoch.CurrentEpochStartTime = now.Add(-80 * 24 * time.Hour) // 80 days ago
		case "airdrop-juno":
			epoch.Duration = time.Hour * 24 * 30                        // 30 days
			epoch.StartTime = now.Add(-110 * 24 * time.Hour)            // 110 days ago - round 2
			epoch.CurrentEpochStartTime = now.Add(-20 * 24 * time.Hour) // 20 days ago
		case "airdrop-osmosis":
			epoch.Duration = time.Hour * 24 * 30                        // 30 days
			epoch.StartTime = now.Add(-140 * 24 * time.Hour)            // 140 days ago - round 3
			epoch.CurrentEpochStartTime = now.Add(-20 * 24 * time.Hour) // 20 days ago
		}
		suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epoch)
	}

	// Get claim metadata
	claimMetadataList := suite.app.ClaimKeeper.GetClaimMetadata(suite.ctx)
	suite.Require().NotNil(claimMetadataList)
	suite.Require().Len(claimMetadataList, 3)

	// Check the contents of the metadata
	for _, metadata := range claimMetadataList {
		suite.Require().Contains([]string{types.DefaultAirdropIdentifier, "juno", "osmosis"}, metadata.AirdropIdentifier, "airdrop ID")
		suite.Require().Equal(now.Add(10*24*time.Hour), metadata.CurrentRoundEnd, "%s round end time", metadata.AirdropIdentifier) // 10 days from now

		switch metadata.AirdropIdentifier {
		case types.DefaultAirdropIdentifier:
			suite.Require().Equal("1", metadata.CurrentRound, "stride current round")
			suite.Require().Equal(now.Add(-80*24*time.Hour), metadata.CurrentRoundStart, "stride round start time") // 80 days ago
		case "juno":
			suite.Require().Equal("2", metadata.CurrentRound, "juno current round")
			suite.Require().Equal(now.Add(-20*24*time.Hour), metadata.CurrentRoundStart, "juno round start time") // 20 days ago
		case "osmosis":
			suite.Require().Equal("3", metadata.CurrentRound, "osmo current round")
			suite.Require().Equal(now.Add(-20*24*time.Hour), metadata.CurrentRoundStart, "osmo round start time") // 20 days ago
		}
	}
}

type UpdateAirdropTestCase struct {
	airdropId     string
	evmosAddress  string
	strideAddress string
	recordKey     sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupUpdateAirdropAddressChangeTests() UpdateAirdropTestCase {
	suite.SetupTest()

	airdropId := "osmosis"

	evmosAddress := "evmos1wg6vh689gw93umxqquhe3yaqf0h9wt9d4q7550"
	strideAddress := "stride1svy5pga6g2er2wjrcujcrg0efce4pref8dksr9"

	recordKeyString := utils.ConvertAddressToStrideAddress(evmosAddress)
	recordKey := sdk.MustAccAddressFromBech32(recordKeyString)

	claimRecord := types.ClaimRecord{
		Address:           recordKeyString,
		Weight:            sdk.NewDecWithPrec(50, 2), // 50%
		ActionCompleted:   []bool{false, false, false},
		AirdropIdentifier: airdropId,
	}

	err := suite.app.ClaimKeeper.SetClaimRecordsWithWeights(suite.ctx, []types.ClaimRecord{claimRecord})
	suite.Require().NoError(err)

	// Create stride account so that it can claim
	suite.app.AccountKeeper.SetAccount(suite.ctx, &authtypes.BaseAccount{Address: strideAddress})

	return UpdateAirdropTestCase{
		airdropId:     airdropId,
		evmosAddress:  evmosAddress,
		recordKey:     recordKey,
		strideAddress: strideAddress,
	}
}

func (suite *KeeperTestSuite) TestUpdateAirdropAddress() {
	tc := suite.SetupUpdateAirdropAddressChangeTests()

	strideAccAddress := sdk.MustAccAddressFromBech32(tc.strideAddress)
	airdropClaimCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000))

	// verify that the Evmos address is different from the address in the key used to store the claim
	suite.Require().NotEqual(tc.evmosAddress, tc.recordKey.String(), "evmos address should not equal the address key")
	// verify new Evmos address starts with "stride"
	suite.Require().True(strings.HasPrefix(tc.recordKey.String(), "stride"), "evmos address should start with stride")

	// Confirm that the user (using the old key'd address) has claimable tokens
	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, tc.recordKey, tc.airdropId, false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins.String(), airdropClaimCoins.String())

	// verify that we can't yet claim with the stride address (because it hasn't been remapped yet)
	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, strideAccAddress, tc.airdropId, false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.NewCoins(), "stride address should claim 0 coins before update", strideAccAddress)

	claims, err := suite.app.ClaimKeeper.GetClaimStatus(suite.ctx, strideAccAddress)
	suite.Require().NoError(err)
	suite.Require().Empty(claims, "stride address should have 0 claim records before update")

	// verify that we can claim the airdrop with the current airdrop key (which represents the incorrect stride address)
	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, tc.recordKey, tc.airdropId, false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(100_000_000))), "parsed evmos address should be allowed to claim")

	claims, err = suite.app.ClaimKeeper.GetClaimStatus(suite.ctx, tc.recordKey)
	suite.Require().NoError(err)

	properClaims := []types.ClaimStatus{{AirdropIdentifier: tc.airdropId, Claimed: false}}
	suite.Require().Equal(claims, properClaims, "evmos address should have 1 claim record before update")

	// update the stride address so that there's now a correct mapping from evmos -> stride address
	err = suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, tc.recordKey.String(), tc.strideAddress, tc.airdropId)
	suite.Require().NoError(err, "airdrop update address should succeed")

	// verify that the old key CAN NOT claim after the update
	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, tc.recordKey, tc.airdropId, false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.NewCoins(), "evmos address should claim 0 coins after update", tc.recordKey)

	claims, err = suite.app.ClaimKeeper.GetClaimStatus(suite.ctx, tc.recordKey)
	suite.Require().NoError(err)
	suite.Require().Empty(claims, "evmos address should have 0 claim records after update")

	// verify that the stride address CAN claim after the update
	coins, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, strideAccAddress, tc.airdropId, false)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(100_000_000))), "stride address should be allowed to claim after update")

	claims, err = suite.app.ClaimKeeper.GetClaimStatus(suite.ctx, strideAccAddress)
	suite.Require().NoError(err)
	suite.Require().Equal(claims, properClaims, "stride address should have 1 claim record after update")

	// claim with the Stride address
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, strideAccAddress, types.ACTION_FREE, tc.airdropId)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(20_000_000))), "stride address should be allowed to claim after update")

	// verify Stride address can't claim again
	coins, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, strideAccAddress, types.ACTION_FREE, tc.airdropId)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(0))), "can't claim twice after update")

	// verify Stride address balance went up
	strideBalance := suite.app.BankKeeper.GetBalance(suite.ctx, strideAccAddress, "stake")
	suite.Require().Equal(strideBalance.Amount, sdk.NewIntFromUint64(20_000_000), "stride address balance should have increased after claiming")
}

func (suite *KeeperTestSuite) TestUpdateAirdropAddress_AirdropNotFound() {
	tc := suite.SetupUpdateAirdropAddressChangeTests()

	// update the address
	err := suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, tc.evmosAddress, tc.strideAddress, "stride")
	suite.Require().Error(err, "airdrop address update should fail with incorrect airdrop id")
}

func (suite *KeeperTestSuite) TestUpdateAirdropAddress_StrideAddressIncorrect() {
	tc := suite.SetupUpdateAirdropAddressChangeTests()

	// update the address
	incorrectStrideAddress := tc.strideAddress + "a"
	err := suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, tc.evmosAddress, incorrectStrideAddress, tc.airdropId)
	suite.Require().Error(err, "airdrop address update should fail with incorrect stride address")
}

func (suite *KeeperTestSuite) TestUpdateAirdropAddress_HostAddressIncorrect() {
	tc := suite.SetupUpdateAirdropAddressChangeTests()

	// should fail with a clearly wrong host address
	err := suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, "evmostest", tc.strideAddress, tc.airdropId)
	suite.Require().Error(err, "airdrop address update should fail with clearly incorrect host address")

	// should fail if host address is not a stride address
	err = suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, tc.evmosAddress, tc.strideAddress, tc.airdropId)
	suite.Require().Error(err, "airdrop address update should fail with host address in wrong zone")

	// should fail is host address (record key) is slightly incorrect
	recordKeyString := tc.recordKey.String()
	modifiedAddress := recordKeyString[:len(recordKeyString)-1] + "a"
	err = suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, modifiedAddress, tc.strideAddress, tc.airdropId)
	suite.Require().Error(err, "airdrop address update should fail with incorrect host address")

	// should fail is host address is correct but doesn't have a claimrecord
	randomStrideAddress := "stride16qv5wnkwwvd2qj5ttwznmngc09cet8l9zhm2ru"
	err = suite.app.ClaimKeeper.UpdateAirdropAddress(suite.ctx, randomStrideAddress, tc.strideAddress, tc.airdropId)
	suite.Require().Error(err, "airdrop address update should fail with not present host address")
}

func (suite *KeeperTestSuite) TestGetAirdropByChainId() {
	// Store 5 airdrops
	airdrops := []*types.Airdrop{}
	for i := 0; i < 5; i++ {
		airdropId := fmt.Sprintf("airdrop-%d", i)
		chainId := fmt.Sprintf("chain-%d", i)

		airdrops = append(airdrops, &types.Airdrop{
			AirdropIdentifier: airdropId,
			ChainId:           chainId,
		})
	}
	err := suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{Airdrops: airdrops})
	suite.Require().NoError(err, "no error expected when setting airdrops")

	// Lookup each airdrop by chain-id
	for i, expected := range airdrops {
		actual, found := suite.app.ClaimKeeper.GetAirdropByChainId(suite.ctx, expected.ChainId)
		suite.Require().True(found, "should have found airdrop %d", i)
		suite.Require().Equal(expected.AirdropIdentifier, actual.AirdropIdentifier, "airdrop identifier for %d", i)
	}

	// Lookup a non-existent airdrop - it should not be found
	_, found := suite.app.ClaimKeeper.GetAirdropByChainId(suite.ctx, "fake_chain_id")
	suite.Require().False(found, "fake_chain_id should not have been found")
}
