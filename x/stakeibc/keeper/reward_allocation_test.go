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

	auctiontypes "github.com/Stride-Labs/stride/v26/x/auction/types"
	epochtypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
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

	// Send all rewrds to auction module
	s.App.StakeibcKeeper.AuctionOffRewardCollectorBalance(s.Ctx)

	// Check Auction module balance
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, rewardAmount)
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, rewardAmount)

	// Check RewardCollector module balance
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcOsmo, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorBalance_NoRewardsAccrued() {
	s.SetupTestRewardAllocation()

	// balances should be 0 before
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, sdkmath.ZeroInt())

	// With no IBC tokens in the rewards collector account, the auction off rewards function should do nothing
	s.App.StakeibcKeeper.AuctionOffRewardCollectorBalance(s.Ctx)

	// balances should be 0 after
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(stakeibctypes.RewardCollectorName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcAtom, sdkmath.ZeroInt())
	s.checkModuleAccountBalance(auctiontypes.ModuleName, IbcOsmo, sdkmath.ZeroInt())
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
