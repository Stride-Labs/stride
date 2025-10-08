package keeper_test

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/Stride-Labs/stride/v29/app/apptesting"
	"github.com/Stride-Labs/stride/v29/utils"
	claimkeeper "github.com/Stride-Labs/stride/v29/x/claim/keeper"

	"github.com/Stride-Labs/stride/v29/x/claim/types"
	stridevestingtypes "github.com/Stride-Labs/stride/v29/x/claim/vesting/types"
)

// Test functionality for loading allocation data(csv)
func (s *KeeperTestSuite) TestLoadAllocationData() {
	s.SetupTest()
	allocations := `identifier,address,weight
osmosis,osmo1g7yxhuppp5x3yqkah5mw29eqq5s4sv2fp6e2eg,0.5
osmosis,osmo1h4astdfzjhcwahtfrh24qtvndzzh49xvtm69fg,0.3
stride,stride1av5lwh0msnafn04xkhdyk6mrykxthrawy7uf3d,0.7
stride,stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk,0.3
stride,stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk,0.5`

	ok := s.App.ClaimKeeper.LoadAllocationData(s.Ctx, allocations)
	s.Require().True(ok)

	totalWeight, err := s.App.ClaimKeeper.GetTotalWeight(s.Ctx, "osmosis")
	s.Require().NoError(err)
	s.Require().True(totalWeight.Equal(sdkmath.LegacyMustNewDecFromStr("0.8")))

	totalWeight, err = s.App.ClaimKeeper.GetTotalWeight(s.Ctx, "stride")
	s.Require().NoError(err)
	s.Require().True(totalWeight.Equal(sdkmath.LegacyMustNewDecFromStr("1")))

	addr, _ := sdk.AccAddressFromBech32("stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk") // hex(stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk) = hex(osmo1g7yxhuppp5x3yqkah5mw29eqq5s4sv2fp6e2eg)
	claimRecord, err := s.App.ClaimKeeper.GetClaimRecord(s.Ctx, addr, "osmosis")
	s.Require().NoError(err)
	s.Require().Equal(claimRecord.Address, "stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk")
	s.Require().True(claimRecord.Weight.Equal(sdkmath.LegacyMustNewDecFromStr("0.5")))
	s.Require().Equal(claimRecord.ActionCompleted, []bool{false, false, false})

	claimRecord, err = s.App.ClaimKeeper.GetClaimRecord(s.Ctx, addr, "stride")
	s.Require().NoError(err)
	s.Require().True(claimRecord.Weight.Equal(sdkmath.LegacyMustNewDecFromStr("0.3")))
	s.Require().Equal(claimRecord.ActionCompleted, []bool{false, false, false})
}

// Check unclaimable account's balance after staking
func (s *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	s.SetupTest()

	// Set a normal user account
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	s.SetNewAccount(addr1)

	claim, err := s.App.ClaimKeeper.GetClaimRecord(s.Ctx, addr1, "stride")
	s.NoError(err)
	s.Equal(types.ClaimRecord{}, claim)

	s.App.ClaimKeeper.AfterLiquidStake(s.Ctx, addr1)

	// Get balances for the account
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
	s.Equal(sdk.Coins{}, balances)
}

// Check balances before and after airdrop starts
func (s *KeeperTestSuite) TestHookBeforeAirdropStart() {
	s.SetupTest()

	airdropStartTime := time.Now().Add(time.Hour)

	err := s.App.ClaimKeeper.SetParams(s.Ctx, types.Params{
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
	s.Require().NoError(err)

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	val1 := sdk.ValAddress(addr1)

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}
	s.SetNewAccount(addr1)
	err = s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, claimRecords)
	s.Require().NoError(err)

	coins, err := s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr1, "stride", false)
	s.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	s.True(coins.Empty())

	coins, err = s.App.ClaimKeeper.GetClaimableAmountForAction(s.Ctx, addr1, types.ACTION_FREE, "stride", false)
	s.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	s.True(coins.Empty())

	err = s.App.ClaimKeeper.AfterDelegationModified(s.Ctx, addr1, val1)
	s.NoError(err)
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
	// Now, it is before starting air drop, so claim module should not send the balances to the user after swap.
	s.True(balances.Empty())

	s.App.ClaimKeeper.AfterLiquidStake(s.Ctx.WithBlockTime(airdropStartTime), addr1)
	balances = s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user after liquid stake.
	claimableAmountForLiquidStake := sdkmath.LegacyNewDecWithPrec(60, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		RoundInt64() // 60% for liquid stake
	s.Require().Equal(balances.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())
}

// Check original user balances after being converted into stride vesting account
func (s *KeeperTestSuite) TestBalancesAfterAccountConversion() {
	s.SetupTest()

	// set a normal account
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	s.SetNewAccount(addr)

	initialBal := int64(1000)
	err := s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
	s.Require().NoError(err)

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}

	err = s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, claimRecords)
	s.Require().NoError(err)

	// check if original account tokens are not affected after stride vesting
	_, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr, types.ACTION_DELEGATE_STAKE, "stride")
	s.Require().NoError(err)
	claimableAmountForStake := sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000 - initialBal)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake

	coinsBal := s.App.BankKeeper.GetAllBalances(s.Ctx, addr)
	s.Require().Equal(coinsBal.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal+claimableAmountForStake)).String())

	spendableCoinsBal := s.App.BankKeeper.SpendableCoins(s.Ctx, addr)
	s.Require().Equal(spendableCoinsBal.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)).String())
}

// Check original user balances after being converted into stride vesting account
func (s *KeeperTestSuite) TestClaimAccountTypes() {
	s.SetupTest()

	// set a normal account
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	// Base Account can claim
	s.SetNewAccount(addr)

	initialBal := int64(1000)
	initialCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal))
	err := s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr, initialCoins)
	s.Require().NoError(err)

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
	}

	err = s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, claimRecords)
	s.Require().NoError(err)

	// check if original account tokens are not affected after stride vesting
	_, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr, types.ACTION_DELEGATE_STAKE, "stride")
	s.Require().NoError(err)
	claimableAmountForStake := sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000 - initialBal)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake

	coinsBal := s.App.BankKeeper.GetAllBalances(s.Ctx, addr)
	s.Require().Equal(coinsBal.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal+claimableAmountForStake)).String())

	spendableCoinsBal := s.App.BankKeeper.SpendableCoins(s.Ctx, addr)
	s.Require().Equal(spendableCoinsBal.String(), initialCoins.String())

	// Verify the account type has changed to stride vesting account
	acc := s.App.AccountKeeper.GetAccount(s.Ctx, addr)
	_, isVestingAcc := acc.(*stridevestingtypes.StridePeriodicVestingAccount)
	s.Require().True(isVestingAcc)

	// Initialize vesting accounts
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	account := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr2)
	err = s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr2, initialCoins)
	s.Require().NoError(err)
	baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(account.(*authtypes.BaseAccount), initialCoins, 0)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, baseVestingAccount)

	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	account = s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr3)
	err = s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr3, initialCoins)
	s.Require().NoError(err)
	continuousVestingAccount, err := vestingtypes.NewContinuousVestingAccount(account.(*authtypes.BaseAccount), initialCoins, 0, 1)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, continuousVestingAccount)

	addr4 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	account = s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr4)
	err = s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr4, initialCoins)
	s.Require().NoError(err)
	periodicVestingAccount, err := vestingtypes.NewPeriodicVestingAccount(account.(*authtypes.BaseAccount), initialCoins, 1, []vestingtypes.Period{{
		Length: 1,
		Amount: initialCoins,
	}})
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, periodicVestingAccount)

	// Init claim records
	for _, addr := range []sdk.AccAddress{addr2, addr3, addr4} {
		claimRecords := []types.ClaimRecord{
			{
				Address:           addr.String(),
				Weight:            sdkmath.LegacyNewDecWithPrec(50, 2),
				ActionCompleted:   []bool{false, false, false},
				AirdropIdentifier: types.DefaultAirdropIdentifier,
			},
		}
		err = s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, claimRecords)
		s.Require().NoError(err)
	}

	// Try to claim tokens with each account type
	for _, addr := range []sdk.AccAddress{addr2, addr3, addr4} {
		_, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr, types.ACTION_DELEGATE_STAKE, "stride")
		s.Require().ErrorContains(err, "only BaseAccount and StridePeriodicVestingAccount can claim")
	}
}

// Run all airdrop flow
func (s *KeeperTestSuite) TestAirdropFlow() {
	s.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	weight := sdkmath.LegacyNewDecWithPrec(50, 2)

	for _, addr := range []sdk.AccAddress{addr1, addr2, addr3} {
		initialBal := int64(1000)
		err := s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		s.Require().NoError(err)
		err = s.App.BankKeeper.SendCoins(s.Ctx, addr, distributors["stride"], sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		s.Require().NoError(err)
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

	err := s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, claimRecords)
	s.Require().NoError(err)

	coins, err := s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr1, "stride", false)
	s.Require().NoError(err)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50_000_000)).String())

	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr2, "stride", false)
	s.Require().NoError(err)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50_000_000)).String())

	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr3, "stride", false)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.Coins{})

	// get rewards amount for free
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_FREE, "stride")
	s.Require().NoError(err)
	claimableAmountForFree := sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_DELEGATE_STAKE, "stride")
	s.Require().NoError(err)
	claimableAmountForStake := sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance is 90000000, claim 20% for stake
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_LIQUID_STAKE, "stride")
	s.Require().NoError(err)
	claimableAmountForLiquidStake := sdkmath.LegacyNewDecWithPrec(60, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance = 80000000, claim 60% for liquid stake
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)).String())

	// get spendable balances 3 months later
	ctx := s.Ctx.WithBlockTime(time.Now().Add(types.DefaultVestingDurationForDelegateStake))
	coinsSpendable := s.App.BankKeeper.SpendableCoins(ctx, addr1)
	s.Require().Equal(coins.String(), coinsSpendable.String())

	// check if claims don't vest after initial period of 3 months
	s.Ctx = s.Ctx.WithBlockTime(time.Now().Add(types.DefaultVestingInitialPeriod))
	err = s.App.ClaimKeeper.ResetClaimStatus(s.Ctx, "stride")
	s.Require().NoError(err)
	_, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_LIQUID_STAKE, "stride")
	s.Require().NoError(err)
	claimableAmountForLiquidStake2 := sdkmath.LegacyNewDecWithPrec(60, 2).
		Mul(sdkmath.LegacyNewDec(50_000_000)).
		Mul(weight).
		RoundInt64() // remaining balance = 50000000*(60/100), claim 60% for liquid stake

	coins = s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake+claimableAmountForLiquidStake2)).String())
	coinsSpendable = s.App.BankKeeper.SpendableCoins(ctx, addr1)
	s.Require().Equal(coins.String(), coinsSpendable.String())

	// end airdrop
	err = s.App.ClaimKeeper.EndAirdrop(s.Ctx, "stride")
	s.Require().NoError(err)
}

// Run multi chain airdrop flow
func (s *KeeperTestSuite) TestMultiChainAirdropFlow() {
	s.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	for _, addr := range []sdk.AccAddress{addr1, addr2} {
		initialBal := int64(1000)
		err := s.App.BankKeeper.SendCoins(s.Ctx, distributors["stride"], addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		s.Require().NoError(err)
		err = s.App.BankKeeper.SendCoins(s.Ctx, addr, distributors["stride"], sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBal)))
		s.Require().NoError(err)
	}

	claimRecords := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
		{
			Address:           addr1.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(30, 2), // 30%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "osmosis",
		},
	}

	err := s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, claimRecords)
	s.Require().NoError(err)

	coins, err := s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr1, "stride", false)
	s.Require().NoError(err)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)).String())

	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr2, "juno", false)
	s.Require().NoError(err)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	identifiers := s.App.ClaimKeeper.GetAirdropIdentifiersForUser(s.Ctx, addr1)
	s.Require().Equal(identifiers[0], types.DefaultAirdropIdentifier)
	s.Require().Equal(identifiers[1], "osmosis")

	// get rewards amount for free (stride, osmosis addresses)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_FREE, "stride")
	s.Require().NoError(err)

	coins1, err := s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_FREE, "osmosis")
	s.Require().NoError(err)

	claimableAmountForFree := sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000, claim 20% for free
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())
	s.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// get rewards amount for stake (stride, osmosis addresses)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_DELEGATE_STAKE, "stride")
	s.Require().NoError(err)

	coins1, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_DELEGATE_STAKE, "osmosis")
	s.Require().NoError(err)

	claimableAmountForStake := sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())
	s.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake (stride, osmosis addresses)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_LIQUID_STAKE, "stride")
	s.Require().NoError(err)

	coins1, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr1, types.ACTION_LIQUID_STAKE, "osmosis")
	s.Require().NoError(err)

	claimableAmountForLiquidStake := sdkmath.LegacyNewDecWithPrec(60, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())
	s.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, (claimableAmountForFree+claimableAmountForStake+claimableAmountForLiquidStake)*2)).String())

	// Verify that the max claimable amount is unchanged, even after claims
	maxCoins, err := s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr1, "stride", true)
	s.Require().NoError(err)
	s.Require().Equal(maxCoins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)).String())
	claimableCoins, err := s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, addr1, "stride", false)
	s.Require().NoError(err)
	s.Require().Equal(claimableCoins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	// check if stride and osmosis airdrops ended properly
	s.Ctx = s.Ctx.WithBlockHeight(1000)
	s.App.ClaimKeeper.EndBlocker(s.Ctx.WithBlockTime(time.Now().Add(types.DefaultAirdropDuration)))
	// for stride
	weight, err := s.App.ClaimKeeper.GetTotalWeight(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().NoError(err)
	s.Require().Equal(weight, sdkmath.LegacyZeroDec())

	records := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().Equal(0, len(records))

	// for osmosis
	weight, err = s.App.ClaimKeeper.GetTotalWeight(s.Ctx, "osmosis")
	s.Require().NoError(err)
	s.Require().Equal(weight, sdkmath.LegacyZeroDec())

	records = s.App.ClaimKeeper.GetClaimRecords(s.Ctx, "osmosis")
	s.Require().Equal(0, len(records))

	//*********************** End of Stride, Osmosis airdrop *************************

	// claim airdrops for juno users after ending stride airdrop
	// get rewards amount for stake (juno user)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ACTION_DELEGATE_STAKE, "juno")
	s.Require().NoError(err)
	claimableAmountForStake = sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		RoundInt64() // remaining balance is 100000000*(80/100), claim 20% for stake
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake)).String())

	// get rewards amount for liquid stake (juno user)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx.WithBlockTime(time.Now().Add(time.Hour)), addr2, types.ACTION_LIQUID_STAKE, "juno")
	s.Require().NoError(err)
	claimableAmountForLiquidStake = sdkmath.LegacyNewDecWithPrec(60, 2).
		Mul(sdkmath.LegacyNewDec(100_000_000)).
		RoundInt64() // remaining balance = 100000000*(80/100)*(80/100), claim 60% for liquid stake
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForLiquidStake)).String())

	// get balance after all claim
	coins = s.App.BankKeeper.GetAllBalances(s.Ctx, addr2)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForStake+claimableAmountForLiquidStake)).String())

	// after 3 years, juno users should be still able to claim
	s.Ctx = s.Ctx.WithBlockTime(time.Now().Add(types.DefaultAirdropDuration))
	err = s.App.ClaimKeeper.ResetClaimStatus(s.Ctx, "juno")
	s.Require().NoError(err)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, addr2, types.ACTION_FREE, "juno")
	s.Require().NoError(err)

	claimableAmountForFree = sdkmath.LegacyNewDecWithPrec(20, 2).
		Mul(sdkmath.LegacyNewDec(20_000_000)).
		RoundInt64() // remaining balance = 20000000*(20/100), claim 20% for free
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, claimableAmountForFree)).String())

	// after 3 years + 1 hour, juno users shouldn't be able to claim anymore
	s.Ctx = s.Ctx.WithBlockTime(time.Now().Add(time.Hour).Add(types.DefaultAirdropDuration))
	s.App.ClaimKeeper.EndBlocker(s.Ctx)
	err = s.App.ClaimKeeper.ResetClaimStatus(s.Ctx, "juno")
	s.Require().NoError(err)
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx.WithBlockTime(time.Now().Add(time.Hour).Add(types.DefaultAirdropDuration)), addr2, types.ACTION_FREE, "juno")
	s.Require().NoError(err)
	s.Require().Equal(coins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)).String())

	weight, err = s.App.ClaimKeeper.GetTotalWeight(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().NoError(err)
	s.Require().Equal(weight, sdkmath.LegacyZeroDec())

	records = s.App.ClaimKeeper.GetClaimRecords(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().Equal(0, len(records))
	//*********************** End of Juno airdrop *************************
}

func (s *KeeperTestSuite) TestAreAllTrue() {
	s.Require().True(claimkeeper.AreAllTrue([]bool{true, true, true}))
	s.Require().False(claimkeeper.AreAllTrue([]bool{true, false, true}))
	s.Require().False(claimkeeper.AreAllTrue([]bool{false, false, false}))
}

func (s *KeeperTestSuite) TestCurrentAirdropRound() {
	startTime := time.Now().Add(-50 * 24 * time.Hour) // 50 days ago
	round := claimkeeper.CurrentAirdropRound(startTime)
	s.Require().Equal(1, round)

	startTime = time.Now().Add(-100 * 24 * time.Hour) // 100 days ago
	round = claimkeeper.CurrentAirdropRound(startTime)
	s.Require().Equal(2, round)

	startTime = time.Now().Add(-130 * 24 * time.Hour) // 130 days ago
	round = claimkeeper.CurrentAirdropRound(startTime)
	s.Require().Equal(3, round)
}

func (s *KeeperTestSuite) TestGetClaimStatus() {
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
	err := s.App.ClaimKeeper.SetParams(s.Ctx, airdrops)
	s.Require().NoError(err)

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
		err := s.App.ClaimKeeper.SetClaimRecord(s.Ctx, claimRecord)
		s.Require().NoError(err)
	}

	expectedClaimStatus := []types.ClaimStatus{
		{AirdropIdentifier: types.DefaultAirdropIdentifier, Claimed: false},
		{AirdropIdentifier: "juno", Claimed: false},
		{AirdropIdentifier: "osmosis", Claimed: true},
	}

	// Confirm status lines up with expectations
	status, err := s.App.ClaimKeeper.GetClaimStatus(s.Ctx, sdk.MustAccAddressFromBech32(address))
	s.Require().NoError(err, "no error expected when getting claim status")

	s.Require().Equal(len(expectedClaimStatus), len(status), "number of airdrops")
	for i := 0; i < len(expectedClaimStatus); i++ {
		s.Require().Equal(expectedClaimStatus[i].AirdropIdentifier, status[i].AirdropIdentifier, "airdrop ID for %d", i)
		s.Require().Equal(expectedClaimStatus[i].AirdropIdentifier, status[i].AirdropIdentifier, "airdrop claimed for %i", i)
	}
}

func (s *KeeperTestSuite) TestGetClaimMetadata() {
	// Update airdrop epochs timing
	now := time.Now().UTC()
	epochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)
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
		s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epoch)
	}

	// Get claim metadata
	claimMetadataList := s.App.ClaimKeeper.GetClaimMetadata(s.Ctx)
	s.Require().NotNil(claimMetadataList)
	s.Require().Len(claimMetadataList, 3)

	// Check the contents of the metadata
	for _, metadata := range claimMetadataList {
		s.Require().Contains([]string{types.DefaultAirdropIdentifier, "juno", "osmosis"}, metadata.AirdropIdentifier, "airdrop ID")
		s.Require().Equal(now.Add(10*24*time.Hour), metadata.CurrentRoundEnd, "%s round end time", metadata.AirdropIdentifier) // 10 days from now

		switch metadata.AirdropIdentifier {
		case types.DefaultAirdropIdentifier:
			s.Require().Equal("1", metadata.CurrentRound, "stride current round")
			s.Require().Equal(now.Add(-80*24*time.Hour), metadata.CurrentRoundStart, "stride round start time") // 80 days ago
		case "juno":
			s.Require().Equal("2", metadata.CurrentRound, "juno current round")
			s.Require().Equal(now.Add(-20*24*time.Hour), metadata.CurrentRoundStart, "juno round start time") // 20 days ago
		case "osmosis":
			s.Require().Equal("3", metadata.CurrentRound, "osmo current round")
			s.Require().Equal(now.Add(-20*24*time.Hour), metadata.CurrentRoundStart, "osmo round start time") // 20 days ago
		}
	}
}

type UpdateAirdropTestCase struct {
	airdropId     string
	evmosAddress  string
	strideAddress string
	recordKey     sdk.AccAddress
}

func (s *KeeperTestSuite) SetupUpdateAirdropAddressChangeTests() UpdateAirdropTestCase {
	s.SetupTest()

	airdropId := "osmosis"

	evmosAddress := "evmos1wg6vh689gw93umxqquhe3yaqf0h9wt9d4q7550"
	strideAddress := "stride1svy5pga6g2er2wjrcujcrg0efce4pref8dksr9"

	recordKeyString := utils.ConvertAddressToStrideAddress(evmosAddress)
	recordKey := sdk.MustAccAddressFromBech32(recordKeyString)

	claimRecord := types.ClaimRecord{
		Address:           recordKeyString,
		Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
		ActionCompleted:   []bool{false, false, false},
		AirdropIdentifier: airdropId,
	}

	err := s.App.ClaimKeeper.SetClaimRecordsWithWeights(s.Ctx, []types.ClaimRecord{claimRecord})
	s.Require().NoError(err)

	// Create stride account so that it can claim
	s.SetNewAccount(sdk.MustAccAddressFromBech32(strideAddress))

	return UpdateAirdropTestCase{
		airdropId:     airdropId,
		evmosAddress:  evmosAddress,
		recordKey:     recordKey,
		strideAddress: strideAddress,
	}
}

func (s *KeeperTestSuite) TestUpdateAirdropAddress() {
	tc := s.SetupUpdateAirdropAddressChangeTests()

	strideAccAddress := sdk.MustAccAddressFromBech32(tc.strideAddress)
	airdropClaimCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000))

	// verify that the Evmos address is different from the address in the key used to store the claim
	s.Require().NotEqual(tc.evmosAddress, tc.recordKey.String(), "evmos address should not equal the address key")
	// verify new Evmos address starts with "stride"
	s.Require().True(strings.HasPrefix(tc.recordKey.String(), "stride"), "evmos address should start with stride")

	// Confirm that the user (using the old key'd address) has claimable tokens
	coins, err := s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, tc.recordKey, tc.airdropId, false)
	s.Require().NoError(err)
	s.Require().Equal(coins.String(), airdropClaimCoins.String())

	// verify that we can't yet claim with the stride address (because it hasn't been remapped yet)
	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, strideAccAddress, tc.airdropId, false)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.NewCoins(), "stride address should claim 0 coins before update", strideAccAddress)

	claims, err := s.App.ClaimKeeper.GetClaimStatus(s.Ctx, strideAccAddress)
	s.Require().NoError(err)
	s.Require().Empty(claims, "stride address should have 0 claim records before update")

	// verify that we can claim the airdrop with the current airdrop key (which represents the incorrect stride address)
	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, tc.recordKey, tc.airdropId, false)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewIntFromUint64(100_000_000))), "parsed evmos address should be allowed to claim")

	claims, err = s.App.ClaimKeeper.GetClaimStatus(s.Ctx, tc.recordKey)
	s.Require().NoError(err)

	properClaims := []types.ClaimStatus{{AirdropIdentifier: tc.airdropId, Claimed: false}}
	s.Require().Equal(claims, properClaims, "evmos address should have 1 claim record before update")

	// update the stride address so that there's now a correct mapping from evmos -> stride address
	err = s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, tc.recordKey.String(), tc.strideAddress, tc.airdropId)
	s.Require().NoError(err, "airdrop update address should succeed")

	// verify that the old key CAN NOT claim after the update
	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, tc.recordKey, tc.airdropId, false)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.NewCoins(), "evmos address should claim 0 coins after update", tc.recordKey)

	claims, err = s.App.ClaimKeeper.GetClaimStatus(s.Ctx, tc.recordKey)
	s.Require().NoError(err)
	s.Require().Empty(claims, "evmos address should have 0 claim records after update")

	// verify that the stride address CAN claim after the update
	coins, err = s.App.ClaimKeeper.GetUserTotalClaimable(s.Ctx, strideAccAddress, tc.airdropId, false)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewIntFromUint64(100_000_000))), "stride address should be allowed to claim after update")

	claims, err = s.App.ClaimKeeper.GetClaimStatus(s.Ctx, strideAccAddress)
	s.Require().NoError(err)
	s.Require().Equal(claims, properClaims, "stride address should have 1 claim record after update")

	// claim with the Stride address
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, strideAccAddress, types.ACTION_FREE, tc.airdropId)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewIntFromUint64(20_000_000))), "stride address should be allowed to claim after update")

	// verify Stride address can't claim again
	coins, err = s.App.ClaimKeeper.ClaimCoinsForAction(s.Ctx, strideAccAddress, types.ACTION_FREE, tc.airdropId)
	s.Require().NoError(err)
	s.Require().Equal(coins, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewIntFromUint64(0))), "can't claim twice after update")

	// verify Stride address balance went up
	strideBalance := s.App.BankKeeper.GetBalance(s.Ctx, strideAccAddress, "stake")
	s.Require().Equal(strideBalance.Amount, sdkmath.NewIntFromUint64(20_000_000), "stride address balance should have increased after claiming")
}

func (s *KeeperTestSuite) TestUpdateAirdropAddress_AirdropNotFound() {
	tc := s.SetupUpdateAirdropAddressChangeTests()

	// update the address
	err := s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, tc.evmosAddress, tc.strideAddress, "stride")
	s.Require().Error(err, "airdrop address update should fail with incorrect airdrop id")
}

func (s *KeeperTestSuite) TestUpdateAirdropAddress_StrideAddressIncorrect() {
	tc := s.SetupUpdateAirdropAddressChangeTests()

	// update the address
	incorrectStrideAddress := tc.strideAddress + "a"
	err := s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, tc.evmosAddress, incorrectStrideAddress, tc.airdropId)
	s.Require().Error(err, "airdrop address update should fail with incorrect stride address")
}

func (s *KeeperTestSuite) TestUpdateAirdropAddress_HostAddressIncorrect() {
	tc := s.SetupUpdateAirdropAddressChangeTests()

	// should fail with a clearly wrong host address
	err := s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, "evmostest", tc.strideAddress, tc.airdropId)
	s.Require().Error(err, "airdrop address update should fail with clearly incorrect host address")

	// should fail if host address is not a stride address
	err = s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, tc.evmosAddress, tc.strideAddress, tc.airdropId)
	s.Require().Error(err, "airdrop address update should fail with host address in wrong zone")

	// should fail is host address (record key) is slightly incorrect
	recordKeyString := tc.recordKey.String()
	modifiedAddress := recordKeyString[:len(recordKeyString)-1] + "a"
	err = s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, modifiedAddress, tc.strideAddress, tc.airdropId)
	s.Require().Error(err, "airdrop address update should fail with incorrect host address")

	// should fail is host address is correct but doesn't have a claimrecord
	randomStrideAddress := "stride16qv5wnkwwvd2qj5ttwznmngc09cet8l9zhm2ru"
	err = s.App.ClaimKeeper.UpdateAirdropAddress(s.Ctx, randomStrideAddress, tc.strideAddress, tc.airdropId)
	s.Require().Error(err, "airdrop address update should fail with not present host address")
}

func (s *KeeperTestSuite) TestGetAirdropByChainId() {
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
	err := s.App.ClaimKeeper.SetParams(s.Ctx, types.Params{Airdrops: airdrops})
	s.Require().NoError(err, "no error expected when setting airdrops")

	// Lookup each airdrop by chain-id
	for i, expected := range airdrops {
		actual, found := s.App.ClaimKeeper.GetAirdropByChainId(s.Ctx, expected.ChainId)
		s.Require().True(found, "should have found airdrop %d", i)
		s.Require().Equal(expected.AirdropIdentifier, actual.AirdropIdentifier, "airdrop identifier for %d", i)
	}

	// Lookup a non-existent airdrop - it should not be found
	_, found := s.App.ClaimKeeper.GetAirdropByChainId(s.Ctx, "fake_chain_id")
	s.Require().False(found, "fake_chain_id should not have been found")
}
