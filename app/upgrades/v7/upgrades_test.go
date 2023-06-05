package v7_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/app"
	"github.com/Stride-Labs/stride/v9/app/apptesting"
	v7 "github.com/Stride-Labs/stride/v9/app/upgrades/v7"
	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	// This isn't the exact type host zone schema as the one that's will be in the store
	// before the upgrade, but the only thing that matters, for the sake of the test,
	// is that it doesn't have min/max redemption rate as attributes
	oldstakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/migrations/v2/types"
)

var (
	DummyUpgradeHeight            = int64(5)
	JunoChainId                   = "juno-1"
	OsmosisChainId                = "osmosis-1"
	OsmosisUnbondingFrequency     = uint64(3)
	InitialJunoUnbondingFrequency = uint64(4)
	ustrd                         = "ustrd"
)

// The block time here is arbitrary, but it's must start at a time that is not at an even hour
var InitialBlockTime = time.Date(2023, 1, 1, 8, 43, 0, 0, time.UTC) // January 1st 2023 at 8:43 AM
var EpochStartTime = time.Date(2023, 1, 1, 8, 00, 0, 0, time.UTC)   // January 1st 2023 at 8:00 AM
var ExpectedHourEpoch = epochstypes.EpochInfo{
	Identifier:            epochstypes.HOUR_EPOCH,
	Duration:              time.Hour,
	CurrentEpoch:          0,
	StartTime:             EpochStartTime,
	CurrentEpochStartTime: EpochStartTime,
}
var ExpectedJunoUnbondingFrequency = uint64(5)
var ExpectedEpochProvisions = sdk.NewDec(1_078_767_123)
var ExpectedAllowMessages = []string{
	"/cosmos.bank.v1beta1.MsgSend",
	"/cosmos.bank.v1beta1.MsgMultiSend",
	"/cosmos.staking.v1beta1.MsgDelegate",
	"/cosmos.staking.v1beta1.MsgUndelegate",
	"/cosmos.staking.v1beta1.MsgBeginRedelegate",
	"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
	"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
	"/ibc.applications.transfer.v1.MsgTransfer",
	"/cosmos.gov.v1beta1.MsgVote",
	"/stride.stakeibc.MsgLiquidStake",
	"/stride.stakeibc.MsgRedeemStake",
	"/stride.stakeibc.MsgClaimUndelegatedTokens",
}

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// Setup store before upgrade
	s.SetupEpochs()
	s.SetupHostZones()
	s.SetupIncentiveDiversification()

	// Run the upgrade and iterate 1 block
	s.ConfirmUpgradeSucceededs("v7", DummyUpgradeHeight)

	// Confirm state after upgrade
	s.CheckEpochsAfterUpgrade(true)
	s.CheckInflationAfterUpgrade()
	s.CheckICAAllowMessagesAfterUpgrade()
	s.CheckRedemptionRateSafetyParamsAfterUpgrade()
	s.CheckUnbondingFrequencyAfterUpgrade()
	s.CheckIncentiveDiversificationAfterUpgrade()
	s.CheckRewardCollectorModuleAccountAfterUpgrade()
}

// Sets up the epoch info before the upgrade by adding only the stride and day epoch to the store
func (s *UpgradeTestSuite) SetupEpochs() {
	// Remove any epochs that are initialized by default
	for _, epoch := range s.App.EpochsKeeper.AllEpochInfos(s.Ctx) {
		s.App.EpochsKeeper.DeleteEpochInfo(s.Ctx, epoch.Identifier)
	}

	// Add stride and day epochs
	s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochstypes.EpochInfo{
		Identifier: epochstypes.DAY_EPOCH,
	})
	s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochstypes.EpochInfo{
		Identifier: epochstypes.STRIDE_EPOCH,
	})

	// Change the context to be a time that's not rounded on the hour
	s.Ctx = s.Ctx.WithBlockTime(InitialBlockTime)
}

// Checks that the hour epoch has been added
// For the unit test that calls the AddHourEpoch function directly, the epoch should not have started yet
// But for the full upgrade unit test case, a block will be incremented which should start the epoch
func (s *UpgradeTestSuite) CheckEpochsAfterUpgrade(epochStarted bool) {
	// Confirm stride and day epoch are still present
	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)
	s.Require().Len(allEpochs, 3, "total number of epochs")

	_, found := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochstypes.DAY_EPOCH)
	s.Require().True(found, "day epoch found")
	_, found = s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochstypes.STRIDE_EPOCH)
	s.Require().True(found, "stride epoch found")

	// If the upgrade passed an a block was incremented, the epoch should be started
	expectedHourEpoch := ExpectedHourEpoch
	if epochStarted {
		expectedHourEpoch.CurrentEpoch = 1
		expectedHourEpoch.EpochCountingStarted = true
		expectedHourEpoch.CurrentEpochStartHeight = DummyUpgradeHeight
	} else {
		expectedHourEpoch.EpochCountingStarted = false
		expectedHourEpoch.CurrentEpochStartHeight = s.Ctx.BlockHeight()
	}

	actualHourEpoch, found := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochstypes.HOUR_EPOCH)
	s.Require().True(found, "hour epoch should have been found")
	s.Require().Equal(expectedHourEpoch, actualHourEpoch, "hour epoch info")
}

// Confirms the inflation has been set properly after the upgrade
func (s *UpgradeTestSuite) CheckInflationAfterUpgrade() {
	minter := s.App.MintKeeper.GetMinter(s.Ctx)
	s.Require().Equal(ExpectedEpochProvisions, minter.EpochProvisions)
}

// Confirms the ICA allow messages have been set after the upgrade
func (s *UpgradeTestSuite) CheckICAAllowMessagesAfterUpgrade() {
	params := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(params.HostEnabled, "host enabled")
	s.Require().ElementsMatch(ExpectedAllowMessages, params.AllowMessages, "allow messages")
}

// Stores an old osmo and juno host zone
// Juno should have an unbonding frequency of 4 in the old store
func (s *UpgradeTestSuite) SetupHostZones() {
	codec := app.MakeEncodingConfig().Marshaler
	stakeibcStore := s.Ctx.KVStore(s.App.GetKey(stakeibctypes.StoreKey))
	hostzoneStore := prefix.NewStore(stakeibcStore, stakeibctypes.KeyPrefix(stakeibctypes.HostZoneKey))

	// Redemption rates is required so invariant is not broken during upgrade
	osmosis := oldstakeibctypes.HostZone{
		ChainId:            OsmosisChainId,
		UnbondingFrequency: OsmosisUnbondingFrequency,
		RedemptionRate:     sdk.NewDec(1),
	}
	juno := oldstakeibctypes.HostZone{
		ChainId:            JunoChainId,
		UnbondingFrequency: InitialJunoUnbondingFrequency,
		RedemptionRate:     sdk.NewDec(1),
	}

	osmosisBz, err := codec.Marshal(&osmosis)
	s.Require().NoError(err, "no error expected when marshalling osmosis host zone")
	junoBz, err := codec.Marshal(&juno)
	s.Require().NoError(err, "no error expected when marshalling juno host zone")

	hostzoneStore.Set([]byte(osmosis.ChainId), osmosisBz)
	hostzoneStore.Set([]byte(juno.ChainId), junoBz)
}

// Check that the juno unbondinng frequency was changed after the upgrade
func (s *UpgradeTestSuite) CheckUnbondingFrequencyAfterUpgrade() {
	osmosis, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, OsmosisChainId)
	s.Require().True(found, "osmosis host zone should have been found")
	s.Require().Equal(OsmosisUnbondingFrequency, osmosis.UnbondingFrequency)

	juno, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, JunoChainId)
	s.Require().True(found, "juno host zone should have been found")
	s.Require().Equal(ExpectedJunoUnbondingFrequency, juno.UnbondingFrequency)
}

// Checks that the redemption rate safety params have been set after the upgrade
func (s *UpgradeTestSuite) CheckRedemptionRateSafetyParamsAfterUpgrade() {
	// Confirm the parameters on each host
	allHostZones := s.App.StakeibcKeeper.GetAllHostZone(s.Ctx)
	s.Require().Len(allHostZones, 2, "number of host zones")

	for _, hostZone := range allHostZones {
		s.Require().False(hostZone.Halted, "host zone %s should not be halted", hostZone.ChainId)

		s.Require().Equal(hostZone.MinRedemptionRate, sdk.MustNewDecFromStr("0.9"),
			"host zone %s min redemption rate", hostZone.ChainId)
		s.Require().Equal(hostZone.MaxRedemptionRate, sdk.MustNewDecFromStr("1.5"),
			"host zone %s max redemption rate", hostZone.ChainId)
	}
}

// Funds the relevant accounts for incentive diversification
func (s *UpgradeTestSuite) SetupIncentiveDiversification() {
	// Get addresses for source and destination
	incentiveProgramAddress, err := sdk.AccAddressFromBech32(v7.IncentiveProgramAddress)
	s.Require().NoError(err, "no error expected when converting Incentive Program address")
	strideFoundationAddress, err := sdk.AccAddressFromBech32(v7.StrideFoundationAddress_F4)
	s.Require().NoError(err, "no error expected when converting Stride Foundation address")

	// Fund incentive program account with 23M, and stride foundation with 4.1M
	// (any values can be used here for the test, but these are used to resemble mainnet)
	initialProgram := sdk.NewCoin(ustrd, sdk.NewInt(23_000_000_000_000))
	initialFoundation := sdk.NewCoin(ustrd, sdk.NewInt(4_157_085_999_543))
	s.FundAccount(incentiveProgramAddress, initialProgram)
	s.FundAccount(strideFoundationAddress, initialFoundation)
}

// Check that the incentive diversification transfer was successful
func (s *UpgradeTestSuite) CheckIncentiveDiversificationAfterUpgrade() {
	// Get addresses for source and destination
	incentiveProgramAddress, err := sdk.AccAddressFromBech32(v7.IncentiveProgramAddress)
	s.Require().NoError(err, "no error expected when converting Incentive Program address")
	strideFoundationAddress, err := sdk.AccAddressFromBech32(v7.StrideFoundationAddress_F4)
	s.Require().NoError(err, "no error expected when converting Stride Foundation address")

	// Confirm 3M were sent from the incentive program accoun to the stride foundation
	expectedIncentiveBalance := sdk.NewCoin(ustrd, sdk.NewInt(20_000_000_000_000))
	expectedFoundationBalance := sdk.NewCoin(ustrd, sdk.NewInt(7_157_085_999_543))
	actualIncentiveBalance := s.App.BankKeeper.GetBalance(s.Ctx, incentiveProgramAddress, ustrd)
	actualFoundationBalance := s.App.BankKeeper.GetBalance(s.Ctx, strideFoundationAddress, ustrd)

	s.CompareCoins(expectedIncentiveBalance, actualIncentiveBalance, "incentive balance after upgrade")
	s.CompareCoins(expectedFoundationBalance, actualFoundationBalance, "foundation balance after upgrade")
}

// Check that the reward collector module account has been created after the upgrade
func (s *UpgradeTestSuite) CheckRewardCollectorModuleAccountAfterUpgrade() {
	s.Require().NotNil(s.App.AccountKeeper.GetModuleAddress(stakeibctypes.RewardCollectorName))
}

func (s *UpgradeTestSuite) TestAddHourEpoch() {
	s.SetupEpochs()
	v7.AddHourEpoch(s.Ctx, s.App.EpochsKeeper)
	s.CheckEpochsAfterUpgrade(false)
}

func (s *UpgradeTestSuite) TestIncreaseStrideInflation() {
	v7.IncreaseStrideInflation(s.Ctx, s.App.MintKeeper)
	s.CheckInflationAfterUpgrade()
}

func (s *UpgradeTestSuite) TestAddICAHostAllowMessages() {
	v7.AddICAHostAllowMessages(s.Ctx, s.App.ICAHostKeeper)
	s.CheckICAAllowMessagesAfterUpgrade()
}

func (s *UpgradeTestSuite) TestModifyJunoUnbondingFrequency() {
	s.SetupHostZones()

	err := v7.ModifyJunoUnbondingFrequency(s.Ctx, s.App.StakeibcKeeper)
	s.Require().NoError(err)

	s.CheckUnbondingFrequencyAfterUpgrade()
}

func (s *UpgradeTestSuite) TestAddRedemptionRateSafetyChecks() {
	s.SetupHostZones()
	v7.AddRedemptionRateSafetyChecks(s.Ctx, s.App.StakeibcKeeper)
	s.CheckRedemptionRateSafetyParamsAfterUpgrade()
}

func (s *UpgradeTestSuite) TestIncentiveDiversification() {
	s.SetupIncentiveDiversification()

	err := v7.ExecuteProp153(s.Ctx, s.App.BankKeeper)
	s.Require().NoError(err)

	s.CheckIncentiveDiversificationAfterUpgrade()
}

func (s *UpgradeTestSuite) TestCreateRewardCollectorModuleAccount() {
	err := v7.CreateRewardCollectorModuleAccount(s.Ctx, s.App.AccountKeeper)
	s.Require().NoError(err)
	s.CheckRewardCollectorModuleAccountAfterUpgrade()
}
