package v7_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v6/app"
	"github.com/Stride-Labs/stride/v6/app/apptesting"
	v7 "github.com/Stride-Labs/stride/v6/app/upgrades/v7"
	epochstypes "github.com/Stride-Labs/stride/v6/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"

	// This isn't the exact type host zone schema as the one that's will be in the store
	// before the upgrade, but the only thing that matters, for the sake of the test,
	// is that it doesn't have min/max redemption rate as attributes
	oldstakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/migrations/v2/types"
)

var (
	DummyUpgradeHeight = int64(5)
	JunoChainId        = "juno-1"
	OsmosisChainId     = "osmosis-1"
	ustrd              = "ustrd"
)
var ExpectedHourEpoch = epochstypes.EpochInfo{
	Identifier:              epochstypes.HOUR_EPOCH,
	StartTime:               time.Time{},
	Duration:                time.Hour,
	CurrentEpoch:            0,
	CurrentEpochStartHeight: 0,
	CurrentEpochStartTime:   time.Time{},
	EpochCountingStarted:    false,
}
var ExpectedJunoUnbondingFrequency = uint64(5)
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
	s.SetupUpgrade()
	s.ConfirmUpgradeSucceededs("v7", DummyUpgradeHeight)
	s.CheckStateAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupUpgrade() {
	codec := app.MakeEncodingConfig().Marshaler

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

	// Add host zones (use old types so that the min/max redemption rate are not present)
	stakeibcStore := s.Ctx.KVStore(s.App.GetKey(stakeibctypes.StoreKey))
	hostzoneStore := prefix.NewStore(stakeibcStore, stakeibctypes.KeyPrefix(stakeibctypes.HostZoneKey))

	// Redemption rates is required so invariant is not broken during upgrade
	osmosis := oldstakeibctypes.HostZone{
		ChainId:            OsmosisChainId,
		UnbondingFrequency: 3,
		RedemptionRate:     sdk.NewDec(1),
	}
	juno := oldstakeibctypes.HostZone{
		ChainId:            JunoChainId,
		UnbondingFrequency: 4,
		RedemptionRate:     sdk.NewDec(1),
	}
	osmosisBz, err := codec.Marshal(&osmosis)
	s.Require().NoError(err, "no error expected when marshalling osmosis host zone")
	junoBz, err := codec.Marshal(&juno)
	s.Require().NoError(err, "no error expected when marshalling juno host zone")

	hostzoneStore.Set([]byte(osmosis.ChainId), osmosisBz)
	hostzoneStore.Set([]byte(juno.ChainId), junoBz)

	// Get addresses for source and destination
	incentiveProgramAddress, err := sdk.AccAddressFromBech32(v7.IncentiveProgramAddress)
	s.Require().NoError(err, "no error expected when converting Incentive Program address")
	strideFoundationAddress, err := sdk.AccAddressFromBech32(v7.StrideFoundationAddress)
	s.Require().NoError(err, "no error expected when converting Stride Foundation address")

	// Fund incentive program account with 23M, and stride foundation with 4.1M
	// (any values can be used here for the test, but these are used to resemble mainnet)
	initialProgram := sdk.NewCoin(ustrd, sdk.NewInt(23_000_000_000_000))
	initialFoundation := sdk.NewCoin(ustrd, sdk.NewInt(4_157_085_999_543))
	s.FundAccount(incentiveProgramAddress, initialProgram)
	s.FundAccount(strideFoundationAddress, initialFoundation)
}

func (s *UpgradeTestSuite) CheckStateAfterUpgrade() {
	// Check epochs
	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)
	s.Require().Len(allEpochs, 3, "total number of epochs")

	_, found := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochstypes.DAY_EPOCH)
	s.Require().True(found, "day epoch found")
	_, found = s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochstypes.STRIDE_EPOCH)
	s.Require().True(found, "stride epoch found")

	// Epoch should have been started after the upgrade
	hourEpochAfterUpgrade := ExpectedHourEpoch
	hourEpochAfterUpgrade.EpochCountingStarted = true
	hourEpochAfterUpgrade.CurrentEpoch = 1
	hourEpochAfterUpgrade.CurrentEpochStartHeight = DummyUpgradeHeight

	actualHourEpoch, found := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochstypes.HOUR_EPOCH)
	s.Require().True(found)
	s.Require().Equal(hourEpochAfterUpgrade, actualHourEpoch, "hour epoch found")

	// Check allow messages
	params := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(params.HostEnabled, "host enabled")
	s.Require().ElementsMatch(ExpectedAllowMessages, params.AllowMessages, "allow messages")

	// Check host zones
	allHostZones := s.App.StakeibcKeeper.GetAllHostZone(s.Ctx)
	s.Require().Len(allHostZones, 2, "total number of host zones")

	_, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, "osmosis-1")
	s.Require().True(found, "osmosis host zone should have been found")

	juno, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, JunoChainId)
	s.Require().True(found, "juno host zone should have been found")

	s.Require().Equal(ExpectedJunoUnbondingFrequency, juno.UnbondingFrequency)

	// Confirm balances after incentive diversification
	incentiveProgramAddress, err := sdk.AccAddressFromBech32(v7.IncentiveProgramAddress)
	s.Require().NoError(err, "no error expected when converting Incentive Program address")
	strideFoundationAddress, err := sdk.AccAddressFromBech32(v7.StrideFoundationAddress)
	s.Require().NoError(err, "no error expected when converting Stride Foundation address")

	expectedIncentiveBalance := sdk.NewCoin(ustrd, sdk.NewInt(20_000_000_000_000))
	expectedFoundationBalance := sdk.NewCoin(ustrd, sdk.NewInt(7_157_085_999_543))
	actualIncentiveBalance := s.App.BankKeeper.GetBalance(s.Ctx, incentiveProgramAddress, ustrd)
	actualFoundationBalance := s.App.BankKeeper.GetBalance(s.Ctx, strideFoundationAddress, ustrd)

	s.CompareCoins(expectedIncentiveBalance, actualIncentiveBalance, "incentive balance after upgrade")
	s.CompareCoins(expectedFoundationBalance, actualFoundationBalance, "foundation balance after upgrade")
}

func (s *UpgradeTestSuite) TestAddHourEpoch() {
	v7.AddHourEpoch(s.Ctx, s.App.EpochsKeeper)

	actualEpochInfo, found := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "hour")
	s.Require().True(found, "hour epoch should have been found")
	s.Require().Equal(ExpectedHourEpoch, actualEpochInfo, "epoch info")
}

func (s *UpgradeTestSuite) TestAddICAHostAllowMessages() {
	v7.AddICAHostAllowMessages(s.Ctx, s.App.ICAHostKeeper)

	params := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(params.HostEnabled, "host enabled")
	s.Require().ElementsMatch(ExpectedAllowMessages, params.AllowMessages, "allow messages")
}

func (s *UpgradeTestSuite) TestModifyJunoUnbondingFrequency() {
	osmoUnbondingFrequency := uint64(3)
	initialJunoUnbondingFrequency := uint64(4)

	// Add osmo and juno host zones
	hostZones := []stakeibctypes.HostZone{
		{
			ChainId:            OsmosisChainId,
			UnbondingFrequency: osmoUnbondingFrequency,
		},
		{
			ChainId:            JunoChainId,
			UnbondingFrequency: initialJunoUnbondingFrequency,
		},
	}
	for _, hz := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)
	}

	// Modify the juno unbonding frequency
	v7.ModifyJunoUnbondingFrequency(s.Ctx, s.App.StakeibcKeeper)

	// Confirm the osmo and juno unbonding frequencies
	osmosis, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, OsmosisChainId)
	s.Require().True(found, "osmosis host zone should have been found")
	s.Require().Equal(osmoUnbondingFrequency, osmosis.UnbondingFrequency)

	juno, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, JunoChainId)
	s.Require().True(found, "juno host zone should have been found")
	s.Require().Equal(ExpectedJunoUnbondingFrequency, juno.UnbondingFrequency)
}

func (s *UpgradeTestSuite) TestAddMinMaxRedemptionRate() {

}

func (s *UpgradeTestSuite) TestIncentiveDiversification() {
	// Get addresses for source and destination
	incentiveProgramAddress, err := sdk.AccAddressFromBech32(v7.IncentiveProgramAddress)
	s.Require().NoError(err, "no error expected when converting Incentive Program address")
	strideFoundationAddress, err := sdk.AccAddressFromBech32(v7.StrideFoundationAddress)
	s.Require().NoError(err, "no error expected when converting Stride Foundation address")

	// Fund incentive program account with 23M, and stride foundation with 4.1M
	// (any values can be used here for the test, but these are used to resemble mainnet)
	initialProgram := sdk.NewCoin(ustrd, sdk.NewInt(23_000_000_000_000))
	initialFoundation := sdk.NewCoin(ustrd, sdk.NewInt(4_157_085_999_543))
	s.FundAccount(incentiveProgramAddress, initialProgram)
	s.FundAccount(strideFoundationAddress, initialFoundation)

	// Trigger bank send from upgrade
	v7.IncentiveDiversification(s.Ctx, s.App.BankKeeper)

	// Confirm balances
	expectedIncentiveBalance := sdk.NewCoin(ustrd, sdk.NewInt(20_000_000_000_000))
	expectedFoundationBalance := sdk.NewCoin(ustrd, sdk.NewInt(7_157_085_999_543))
	actualIncentiveBalance := s.App.BankKeeper.GetBalance(s.Ctx, incentiveProgramAddress, ustrd)
	actualFoundationBalance := s.App.BankKeeper.GetBalance(s.Ctx, strideFoundationAddress, ustrd)

	s.CompareCoins(expectedIncentiveBalance, actualIncentiveBalance, "incentive balance after upgrade")
	s.CompareCoins(expectedFoundationBalance, actualFoundationBalance, "foundation balance after upgrade")
}
