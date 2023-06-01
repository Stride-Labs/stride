package keeper_test

import (
	"strings"

	sdkmath "cosmossdk.io/math"

	abci "github.com/cometbft/cometbft/abci/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	teststaking "github.com/cosmos/cosmos-sdk/x/staking/testutil"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupTestRewardAllocation() {
	// Create two host zones so we can map the ibc and st denom's back to a host zone
	// We need valid addresses for the module account addresses, otherwise liquid stake will fail
	hostZone1 := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.OneDec(),
		Address:        stakeibctypes.NewZoneAddress(HostChainId).String(),
	}
	hostZone2 := stakeibctypes.HostZone{
		ChainId:        OsmoChainId,
		HostDenom:      Osmo,
		IbcDenom:       IbcOsmo,
		RedemptionRate: sdk.OneDec(),
		Address:        stakeibctypes.NewZoneAddress(OsmoChainId).String(),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone2)

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

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_Success() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(1000)

	// Fund reward collector account with ibc'd reward tokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcAtom, rewardAmount))
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcOsmo, rewardAmount))

	// Liquid stake all hostzone token then get stTokens back
	rewardsAccrued := s.App.StakeibcKeeper.LiquidStakeRewardCollectorBalance(s.Ctx, s.GetMsgServer())
	s.Require().True(rewardsAccrued, "rewards should have been liquid staked")

	// Reward Collector acct should have all ibc/XXX tokens converted to stTokens
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StAtom, rewardAmount)
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StOsmo, rewardAmount)

	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_NoRewardsAccrued() {
	s.SetupTestRewardAllocation()

	// With no IBC tokens in the rewards collector account, the liquid stake rewards function should return false
	rewardsAccrued := s.App.StakeibcKeeper.LiquidStakeRewardCollectorBalance(s.Ctx, s.GetMsgServer())
	s.Require().False(rewardsAccrued, "no rewards should have been liquid staked")

	// There should also be no stTokens in the account
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_BalanceDoesNotBelongToHost() {
	s.SetupTestRewardAllocation()
	amount := sdkmath.NewInt(1000)

	// Fund the reward collector with ibc/atom and a denom that is not registerd to a host zone
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(IbcAtom, amount))
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin("fake_denom", amount))

	// Liquid stake should only succeed with atom
	rewardsAccrued := s.App.StakeibcKeeper.LiquidStakeRewardCollectorBalance(s.Ctx, s.GetMsgServer())
	s.Require().True(rewardsAccrued, "rewards should have been liquid staked")

	// The atom should have been liquid staked
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StAtom, amount)

	// But the fake denom and uosmo should not have been touched
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, "fake_denom", amount)
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestSweepRewardCollToFeeCollector_Success() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(1000)

	// Add stTokens to reward collector
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(StAtom, rewardAmount))
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(StOsmo, rewardAmount))

	// Sweep stTokens from Reward Collector to Fee Collector
	err := s.App.StakeibcKeeper.SweepStTokensFromRewardCollToFeeColl(s.Ctx)
	s.Require().NoError(err)

	// Fee Collector acct should have stTokens after they're swept there from Reward Collector
	// The reward collector should have nothing
	s.checkModuleAccountBalance(authtypes.FeeCollectorName, StAtom, rewardAmount)
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StAtom, sdkmath.ZeroInt())

	s.checkModuleAccountBalance(authtypes.FeeCollectorName, StOsmo, rewardAmount)
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, StOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestSweepRewardCollToFeeCollector_NonStTokens() {
	s.SetupTestRewardAllocation()
	amount := sdkmath.NewInt(1000)
	nonStTokenDenom := "XXX"

	// Fund reward collector account with stTokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(nonStTokenDenom, amount))

	// Sweep stTokens from Reward Collector to Fee Collector
	err := s.App.StakeibcKeeper.SweepStTokensFromRewardCollToFeeColl(s.Ctx)
	s.Require().NoError(err)

	// Reward Collector acct should still contain nonStTokenDenom after stTokens after they're swept
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, nonStTokenDenom, amount)

	// Fee Collector acct should have nothing
	s.checkModuleAccountBalance(authtypes.FeeCollectorName, nonStTokenDenom, sdkmath.ZeroInt())
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
		s.FundAccount(acc, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	}
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(s.T(), s.Ctx, &s.App.StakingKeeper)

	pubkeys := simtestutil.CreateTestPubKeys(2)
	stakeAmount := sdk.NewInt(100)

	// create validator with 50% commission
	commission := sdk.NewDecWithPrec(5, 1)
	tstaking.Commission = stakingtypes.NewCommissionRates(commission, commission, sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], pubkeys[0], stakeAmount, true)

	// create second validator with 0% commission
	commission = sdk.NewDec(0)
	tstaking.Commission = stakingtypes.NewCommissionRates(commission, commission, sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], pubkeys[1], stakeAmount, true)

	s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{})
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
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}
	s.App.DistrKeeper.AllocateTokens(s.Ctx, 200, votes)

	// Withdraw rewards
	rewards1, err := s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	s.Require().NoError(err, "no error expected with withdrawing delegator rewards")

	rewards2, err := s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[1]), valAddrs[1])
	s.Require().NoError(err, "no error expected with withdrawing delegator rewards")

	// Check balances contains stTokens
	s.Require().True(strings.Contains(rewards1.String(), "stuatom"))
	s.Require().True(strings.Contains(rewards2.String(), "stuatom"))
}
