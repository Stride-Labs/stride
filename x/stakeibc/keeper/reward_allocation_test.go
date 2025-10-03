package keeper_test

import (
	"strings"

	sdkmath "cosmossdk.io/math"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	teststaking "github.com/cosmos/cosmos-sdk/x/staking/testutil"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/utils"
	auctiontypes "github.com/Stride-Labs/stride/v28/x/auction/types"
	epochtypes "github.com/Stride-Labs/stride/v28/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v28/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupTestRewardAllocation() {
	// Create two host zones so we can map the ibc and st denom's back to a host zone
	// We need valid addresses for the module account addresses, otherwise liquid stake will fail
	hostZone1 := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdkmath.LegacyOneDec(),
		DepositAddress: stakeibctypes.NewHostZoneDepositAddress(HostChainId).String(),
	}
	hostZone2 := stakeibctypes.HostZone{
		ChainId:        OsmoChainId,
		HostDenom:      Osmo,
		IbcDenom:       IbcOsmo,
		RedemptionRate: sdkmath.LegacyOneDec(),
		DepositAddress: stakeibctypes.NewHostZoneDepositAddress(OsmoChainId).String(),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone2)

	// ConsumerRedistributionFraction = how much Stride keeps
	// Set consumer redistribution fraction to 0.85 (same as mainnet)
	consumerParams := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	consumerParams.ConsumerRedistributionFraction = "0.85"
	s.App.ConsumerKeeper.SetParams(s.Ctx, consumerParams)

	// Set epoch tracker and deposit records for liquid stake
	currentEpoch := uint64(2)
	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		NextEpochStartTime: uint64(10),
	}
	initialDepositRecord1 := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: currentEpoch,
		HostZoneId:         HostChainId,
		Amount:             sdkmath.ZeroInt(),
	}
	initialDepositRecord2 := recordtypes.DepositRecord{
		Id:                 2,
		DepositEpochNumber: currentEpoch,
		HostZoneId:         OsmoChainId,
		Amount:             sdkmath.ZeroInt(),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord1)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord2)
}

// Helper function to check the balance of a module account
func (s *KeeperTestSuite) checkModuleAccountBalance(moduleName, denom string, expectedBalance sdkmath.Int) {
	address := s.App.AccountKeeper.GetModuleAccount(s.Ctx, moduleName).GetAddress()
	tokens := s.App.BankKeeper.GetBalance(s.Ctx, address, denom)
	s.Require().Equal(expectedBalance.Int64(), tokens.Amount.Int64(), "%s %s balance", moduleName, denom)
}

// Helper function to check the balance of a regular account
func (s *KeeperTestSuite) checkAccountBalance(address string, denom string, expectedBalance sdkmath.Int) {
	tokens := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(address), denom)
	s.Require().Equal(expectedBalance.Int64(), tokens.Amount.Int64(), "%s %s balance", address, denom)
}

// Helper function to get total stToken balance across all PoA validators
func (s *KeeperTestSuite) getTotalPoAValidatorStTokenBalance(denom string) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, validator := range utils.PoaValidatorSet {
		balance := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(validator), denom)
		total = total.Add(balance.Amount)
	}
	return total
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_Success() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(1000)

	// Fund reward collector account with ibc'd reward tokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcAtom, rewardAmount))
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcOsmo, rewardAmount))

	// Distribute rewards using new logic: 15% liquid staked to PoA validators, 85% to auction
	s.App.StakeibcKeeper.AuctionOffRewardCollectorBalance(s.Ctx)

	// Check PoA validators received stTokens (15% of original amount gets liquid staked)
	// Since the redemption rate is 1:1, 15% of 1000 = 150 stTokens total
	// This gets distributed equally among PoA validators, so 150 / 7 = 21.42 -> 21 per validator
	numValidators := sdkmath.NewInt(7)
	expectedStTokenPerValidator := sdkmath.NewInt(21)
	expectedTotalStTokens := expectedStTokenPerValidator.Mul(numValidators) // adjusts for ignored remainder

	actualStAtomTotal := s.getTotalPoAValidatorStTokenBalance(StAtom)
	actualStOsmoTotal := s.getTotalPoAValidatorStTokenBalance(StOsmo)
	s.Require().Equal(expectedTotalStTokens.Int64(), actualStAtomTotal.Int64(), "total stAtom distributed to PoA validators")
	s.Require().Equal(expectedTotalStTokens.Int64(), actualStOsmoTotal.Int64(), "total stOsmo distributed to PoA validators")

	// Check each validator received equal share (ignoring remainder for simplicity)
	for _, validator := range utils.PoaValidatorSet {
		s.checkAccountBalance(validator, StAtom, expectedStTokenPerValidator)
		s.checkAccountBalance(validator, StOsmo, expectedStTokenPerValidator)
	}

	// Check Auction module balance (should have remainder - 85% of original)
	auctionPortion := sdkmath.NewInt(850) // 85% of 1000
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, auctionPortion)
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, auctionPortion)

	// Check RewardCollector module balance (should be empty)
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_NoRewardsAccrued() {
	s.SetupTestRewardAllocation()

	// Reward account balances should be 0 before
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, sdkmath.ZeroInt())

	// PoA validators should have no stTokens initially
	s.Require().Equal(sdkmath.ZeroInt().Int64(), s.getTotalPoAValidatorStTokenBalance(StAtom).Int64())
	s.Require().Equal(sdkmath.ZeroInt().Int64(), s.getTotalPoAValidatorStTokenBalance(StOsmo).Int64())

	// Auction accounts should have not tokens
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, sdkmath.ZeroInt())

	// With no IBC tokens in the rewards collector account, the distribution function should do nothing
	s.App.StakeibcKeeper.AuctionOffRewardCollectorBalance(s.Ctx)

	// Reward collector balance should be unchanged
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, sdkmath.ZeroInt())

	// PoA validators should be unchanged
	s.Require().Equal(sdkmath.ZeroInt().Int64(), s.getTotalPoAValidatorStTokenBalance(StAtom).Int64())
	s.Require().Equal(sdkmath.ZeroInt().Int64(), s.getTotalPoAValidatorStTokenBalance(StOsmo).Int64())

	// Auction account should be unchanged
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_TotalValidatorShareTruncatesZero() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(6) // truncates validator share to 0.15 * 6 = 0.9 -> 0

	// Fund reward collector account with ibc'd reward tokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcAtom, rewardAmount))
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcOsmo, rewardAmount))

	// Distribute rewards using new logic: 15% liquid staked to PoA validators, 85% to auction
	s.App.StakeibcKeeper.AuctionOffRewardCollectorBalance(s.Ctx)

	// Check PoA validators did not receive anything
	actualStAtomTotal := s.getTotalPoAValidatorStTokenBalance(StAtom)
	actualStOsmoTotal := s.getTotalPoAValidatorStTokenBalance(StOsmo)
	s.Require().Equal(int64(0), actualStAtomTotal.Int64(), "total stAtom distributed to PoA validators")
	s.Require().Equal(int64(0), actualStOsmoTotal.Int64(), "total stOsmo distributed to PoA validators")

	// Check each validator received equal share (ignoring remainder for simplicity)
	for _, validator := range utils.PoaValidatorSet {
		s.checkAccountBalance(validator, StAtom, sdkmath.ZeroInt())
		s.checkAccountBalance(validator, StOsmo, sdkmath.ZeroInt())
	}

	// Auction balance should also be 0 since it short circuits
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, sdkmath.ZeroInt())

	// Check RewardCollector module balance should have the initial amount
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, rewardAmount)
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, rewardAmount)
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_IndividualValidatorShareTruncatesZero() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(10)

	// Fund reward collector account with ibc'd reward tokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcAtom, rewardAmount))
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcOsmo, rewardAmount))

	// Distribute rewards using new logic: 15% liquid staked to PoA validators, 85% to auction
	s.App.StakeibcKeeper.AuctionOffRewardCollectorBalance(s.Ctx)

	// Check PoA validators did not receive anything
	// Total validator share is 0.15 * 10 = 1.5
	// Individual validator share is 1.5 / 7 = 0.2 -> 0
	actualStAtomTotal := s.getTotalPoAValidatorStTokenBalance(StAtom)
	actualStOsmoTotal := s.getTotalPoAValidatorStTokenBalance(StOsmo)
	s.Require().Equal(int64(0), actualStAtomTotal.Int64(), "total stAtom distributed to PoA validators")
	s.Require().Equal(int64(0), actualStOsmoTotal.Int64(), "total stOsmo distributed to PoA validators")

	// Check each validator received equal share (ignoring remainder for simplicity)
	for _, validator := range utils.PoaValidatorSet {
		s.checkAccountBalance(validator, StAtom, sdkmath.ZeroInt())
		s.checkAccountBalance(validator, StOsmo, sdkmath.ZeroInt())
	}

	// Check Auction module balance
	// The 0.15 * 10 = 1 rewards should have been liquid staked, so there should be 9 in the auction
	expectedRewardRemainder := sdkmath.NewInt(9)
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, expectedRewardRemainder)
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, expectedRewardRemainder)

	// The reward collector should have zero remainder
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, sdkmath.ZeroInt())
}

// Test the process of a delegator claiming staking reward stTokens (tests that Fee Account can distribute arbitrary denoms)
func (s *KeeperTestSuite) TestClaimStakingRewardStTokens() {
	s.SetupTestRewardAllocation()
	amount := sdkmath.NewInt(1000)

	// Fund fee collector account with stTokens
	s.FundModuleAccount(authtypes.FeeCollectorName, sdk.NewCoin("st"+Atom, amount))

	// Set up validators & delegators on Stride
	addrs := s.TestAccs
	for _, acc := range addrs {
		s.FundAccount(acc, sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1000000)))
	}
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(s.T(), s.Ctx, &s.App.StakingKeeper)

	pubkeys := simtestutil.CreateTestPubKeys(2)
	stakeAmount := sdkmath.NewInt(100)

	// create validator with 50% commission
	commission := sdkmath.LegacyNewDecWithPrec(5, 1)
	tstaking.Commission = stakingtypes.NewCommissionRates(commission, commission, sdkmath.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], pubkeys[0], stakeAmount, true)

	// create second validator with 0% commission
	commission = sdkmath.LegacyNewDec(0)
	tstaking.Commission = stakingtypes.NewCommissionRates(commission, commission, sdkmath.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[1], pubkeys[1], stakeAmount, true)

	_, err := s.App.EndBlocker(s.Ctx)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)

	// Simulate the token distribution from feeCollector to validators
	abciValA := abci.Validator{
		Address: pubkeys[0].Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: pubkeys[1].Address(),
		Power:   100,
	}
	votes := []abci.VoteInfo{
		{
			Validator:   abciValA,
			BlockIdFlag: cmtproto.BlockIDFlagCommit,
		},
		{
			Validator:   abciValB,
			BlockIdFlag: cmtproto.BlockIDFlagCommit,
		},
	}
	err = s.App.DistrKeeper.AllocateTokens(s.Ctx, 200, votes)
	s.Require().NoError(err)

	// Withdraw rewards
	rewards1, err := s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	s.Require().NoError(err, "no error expected with withdrawing delegator rewards")

	rewards2, err := s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[1]), valAddrs[1])
	s.Require().NoError(err, "no error expected with withdrawing delegator rewards")

	// Check balances contains stTokens
	s.Require().True(strings.Contains(rewards1.String(), "stuatom"))
	s.Require().True(strings.Contains(rewards2.String(), "stuatom"))
}
