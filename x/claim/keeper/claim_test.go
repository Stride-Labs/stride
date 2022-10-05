package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	claim, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.NoError(err)
	suite.Equal(types.ClaimRecord{}, claim)

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(sdk.Coins{}, balances)
}

func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupTest()

	airdropStartTime := time.Now().Add(time.Hour)

	err := suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: time.Hour,
		DurationOfDecay:    time.Hour * 4,
	})
	suite.Require().NoError(err)

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	err = suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	coins, err = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionLiquidStake)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is before starting air drop, so claim module should not send the balances to the user after swap.
	suite.True(balances.Empty())

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx.WithBlockTime(airdropStartTime), addr1)
	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user after swap.
	suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)), balances.AmountOf(sdk.DefaultBondDenom))
}

func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
	suite.SetupTest()

	// airdrop recipient address
	addr1, _ := sdk.AccAddressFromBech32("stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk")

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	params, err := suite.app.ClaimKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	suite.ctx = suite.ctx.WithBlockTime(params.AirdropStartTime.Add(params.DurationUntilDecay).Add(params.DurationOfDecay))

	suite.app.ClaimKeeper.EndAirdrop(suite.ctx)

	suite.Require().NotPanics(func() {
		suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)
	})
}

func (suite *KeeperTestSuite) TestDuplicatedActionNotWithdrawRepeatedly() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins1, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount)

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)
	claim, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.NoError(err)
	suite.True(claim.ActionCompleted[types.ActionLiquidStake])
	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))

	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)
	claim, err = suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.NoError(err)
	suite.True(claim.ActionCompleted[types.ActionLiquidStake])
	claimedCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))
}

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	pub2 := secp256k1.GenPrivKey().PubKey()
	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.Address()), sdk.AccAddress(pub2.Address())}
	claimRecords := []types.ClaimRecord{
		{
			Address:                addrs[0].String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
		{
			Address:                addrs[1].String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}

	// initialize accts
	for i := 0; i < len(addrs); i++ {
		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addrs[i], nil, 0, 0))
	}
	// initialize claim records
	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	// test claim records set
	for i := 0; i < len(addrs); i++ {
		coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[i])
		suite.Require().NoError(err)
		suite.Require().Equal(coins, claimRecords[i].InitialClaimableAmount)
	}

	// set addr[0] as a validator
	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	delAmount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addrs[1],
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, delAmount)))
	suite.Require().NoError(err)
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], delAmount, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	// delegation should automatically call claim and withdraw balance
	actualClaimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[1])
	actualClaimedCoin := actualClaimedCoins.AmountOf(sdk.DefaultBondDenom)
	expectedClaimedCoin := claimRecords[1].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted))))
	suite.Require().Equal(expectedClaimedCoin.String(), actualClaimedCoin.String())

	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], actualClaimedCoin, stakingtypes.Unbonded, validator, true)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}

	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins1, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount, coins1.String())

	coins3, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().Equal(coins3, sdk.Coins{})

	// get rewards amount per action
	coins4, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionDelegateStake)
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)).String())

	// get completed activities
	claimRecord, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.Require().NoError(err)
	for i := range types.Action_name {
		suite.Require().False(claimRecord.ActionCompleted[i])
	}

	// do half of actions
	suite.app.ClaimKeeper.AfterDelegationModified(suite.ctx, addr1, sdk.ValAddress(addr1))

	// check that half are completed
	claimRecord, err = suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(claimRecord.ActionCompleted[types.ActionDelegateStake])

	// get balance after 2 actions done
	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)).String())

	// check that claimable for completed activity is 0
	coins4, err = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionDelegateStake)
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.Coins{}.String())

	// do rest of actions
	suite.app.ClaimKeeper.AfterLiquidStake(suite.ctx, addr1)

	// get balance after rest actions done
	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)).String())

	// get claimable after withdrawing all
	coins1, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(coins1.Empty())

	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx)
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, sdk.DefaultBondDenom)
	suite.Require().Equal(coins, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
}

func (suite *KeeperTestSuite) TestClaimOfDecayed() {
	airdropStartTime := time.Now()
	durationUntilDecay := time.Hour
	durationOfDecay := time.Hour * 4

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}

	tests := []struct {
		fn func()
	}{
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime)
				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionDelegateStake)
				suite.NoError(err)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(2)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

				suite.app.ClaimKeeper.AfterDelegationModified(ctx, addr1, sdk.ValAddress(addr1))
				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(2)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay))
				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionLiquidStake)
				suite.NoError(err)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(2)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

				suite.app.ClaimKeeper.AfterLiquidStake(ctx, addr1)
				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(2)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay / 2))
				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionLiquidStake)
				suite.NoError(err)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

				suite.app.ClaimKeeper.AfterLiquidStake(ctx, addr1)
				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay))
				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionLiquidStake)
				suite.NoError(err)
				suite.True(coins.Empty())

				suite.app.ClaimKeeper.AfterLiquidStake(ctx, addr1)
				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.True(coins.Empty())
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		err := suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
			AirdropStartTime:   airdropStartTime,
			DurationUntilDecay: durationUntilDecay,
			DurationOfDecay:    durationOfDecay,
		})
		suite.NoError(err)

		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
		err = suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
		suite.Require().NoError(err)

		test.fn()
	}
}

func (suite *KeeperTestSuite) TestClawbackAirdrop() {
	suite.SetupTest()

	tests := []struct {
		name           string
		address        string
		sequence       uint64
		expectClawback bool
	}{
		{
			name:           "airdrop address active",
			address:        "stride1g7yxhuppp5x3yqkah5mw29eqq5s4sv2f222xmk",
			sequence:       1,
			expectClawback: false,
		},
		{
			name:           "airdrop address inactive",
			address:        "stride1h4astdfzjhcwahtfrh24qtvndzzh49xvqtfftk",
			sequence:       0,
			expectClawback: true,
		},
		{
			name:           "non airdrop address active",
			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			sequence:       1,
			expectClawback: false,
		},
		{
			name:           "non airdrop address inactive",
			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			sequence:       0,
			expectClawback: false,
		},
	}

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		acc := authtypes.NewBaseAccountWithAddress(addr)
		err = acc.SetSequence(tc.sequence)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		coins := sdk.NewCoins(
			sdk.NewInt64Coin("ustrd", 100))
		simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	}

	err := suite.app.ClaimKeeper.EndAirdrop(suite.ctx)
	suite.Require().NoError(err, "err: %s", err)

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		coins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
		if tc.expectClawback {
			suite.Require().True(coins.IsEqual(sdk.NewCoins()),
				"balance incorrect. test: %s", tc.name)
		} else {
			suite.Require().True(coins.IsEqual(sdk.NewCoins(
				sdk.NewInt64Coin("ustrd", 100),
			)), "balance incorrect. test: %s", tc.name)
		}
	}
}
