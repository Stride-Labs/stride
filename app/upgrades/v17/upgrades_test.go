package v17_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/suite"

	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	icqtypes "github.com/Stride-Labs/stride/v30/x/interchainquery/types"
	recordtypes "github.com/Stride-Labs/stride/v30/x/records/types"

	"github.com/Stride-Labs/stride/v30/app/apptesting"
	v17 "github.com/Stride-Labs/stride/v30/app/upgrades/v17"
	"github.com/Stride-Labs/stride/v30/x/stakeibc/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v30/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v30/x/stakeibc/types"
)

const (
	SommelierChainId = "sommelier-3"

	Atom = "uatom"
	Osmo = "uosmo"
	Somm = "usomm"

	StAtom = "st" + Atom
	StOsmo = "st" + Osmo
	StSomm = "st" + Somm
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinOuterRedemptionRate sdkmath.LegacyDec
	ExpectedMaxOuterRedemptionRate sdkmath.LegacyDec
}

type UpdateRateLimits struct {
	ChainId        string
	ChannelId      string
	RateLimitDenom string
	HostDenom      string
	Duration       uint64
	Threshold      sdkmath.Int
}

type AddRateLimits struct {
	ChannelId    string
	Denom        string
	ChannelValue sdkmath.Int
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
	checkHostZonesAfterUpgrade := s.SetupHostZonesBeforeUpgrade()
	checkMigrateUnbondingRecords := s.SetupMigrateUnbondingRecords()
	checkRateLimitsAfterUpgrade := s.SetupRateLimitsBeforeUpgrade()
	checkCommunityPoolTaxAfterUpgrade := s.SetupCommunityPoolTaxBeforeUpgrade()
	checkQueriesAfterUpgrade := s.SetupQueriesBeforeUpgrade()
	checkProp225AfterUpgrade := s.SetupProp225BeforeUpgrade()

	// Submit upgrade and confirm handler succeeds
	s.ConfirmUpgradeSucceeded(v17.UpgradeName)

	// Check state after upgrade
	checkHostZonesAfterUpgrade()
	checkMigrateUnbondingRecords()
	checkRateLimitsAfterUpgrade()
	checkCommunityPoolTaxAfterUpgrade()
	checkQueriesAfterUpgrade()
	checkProp225AfterUpgrade()
}

// Helper function to check that the community pool stake and redeem holding
// module accounts were registered and stored on the host zone
func (s *UpgradeTestSuite) checkCommunityPoolModuleAccountsRegistered(chainId string) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
	s.Require().True(found, "host zone should have been found")

	s.Require().NotEmpty(hostZone.CommunityPoolStakeHoldingAddress, "stake holding for %s should not be empty", chainId)
	s.Require().NotEmpty(hostZone.CommunityPoolRedeemHoldingAddress, "redeem holding for %s should not be empty", chainId)

	stakeHoldingAddress := sdk.MustAccAddressFromBech32(hostZone.CommunityPoolStakeHoldingAddress)
	redeemHoldingAddress := sdk.MustAccAddressFromBech32(hostZone.CommunityPoolRedeemHoldingAddress)

	stakeHoldingAccount := s.App.AccountKeeper.GetAccount(s.Ctx, stakeHoldingAddress)
	redeemHoldingAccount := s.App.AccountKeeper.GetAccount(s.Ctx, redeemHoldingAddress)

	s.Require().NotNil(stakeHoldingAccount, "stake holding account should have been registered for %s", chainId)
	s.Require().NotNil(redeemHoldingAccount, "redeem holding account should have been registered for %s", chainId)
}

// Helper function to check that the community pool deposit and return ICA accounts were registered for the host zone
// The addresses don't get set until the callback, but we can check that the expected ICA controller port was claimed
// by the ICA controller module
func (s *UpgradeTestSuite) checkCommunityPoolICAAccountsRegistered(chainId string) {
	depositOwner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	returnOwner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_COMMUNITY_POOL_RETURN)

	expectedDepositPortId, _ := icatypes.NewControllerPortID(depositOwner)
	expectedReturnPortId, _ := icatypes.NewControllerPortID(returnOwner)

	_, depositPortIdRegistered := s.App.ScopedICAControllerKeeper.GetCapability(s.Ctx, host.PortPath(expectedDepositPortId))
	_, returnPortIdRegistered := s.App.ScopedICAControllerKeeper.GetCapability(s.Ctx, host.PortPath(expectedReturnPortId))

	s.Require().True(depositPortIdRegistered, "deposit port %s should have been bound", expectedDepositPortId)
	s.Require().True(returnPortIdRegistered, "return port %s should have been bound", expectedReturnPortId)
}

func (s *UpgradeTestSuite) SetupHostZonesBeforeUpgrade() func() {
	hostZones := []stakeibctypes.HostZone{
		{
			ChainId:      v17.GaiaChainId,
			HostDenom:    Atom,
			ConnectionId: ibctesting.FirstConnectionID, // must be connection-0 since an ICA will be submitted
			Validators: []*stakeibctypes.Validator{
				{Address: "val1", SlashQueryInProgress: false},
				{Address: "val2", SlashQueryInProgress: true},
			},
			RedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1"),
		},
		{
			ChainId:      v17.OsmosisChainId,
			HostDenom:    Osmo,
			ConnectionId: "connection-2",
			Validators: []*stakeibctypes.Validator{
				{Address: "val3", SlashQueryInProgress: true},
				{Address: "val4", SlashQueryInProgress: false},
			},
			RedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.2"),
		},
		{
			// This host is just added for the rate limit test
			// No need to validate accounts and redemptino rates
			ChainId:      SommelierChainId,
			HostDenom:    Somm,
			ConnectionId: "connection-3",
			Validators: []*stakeibctypes.Validator{
				{Address: "val5", SlashQueryInProgress: true},
				{Address: "val6", SlashQueryInProgress: false},
			},
			RedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.0"),
		},
	}

	for i, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

		clientId := fmt.Sprintf("07-tendermint-%d", i)
		s.MockClientAndConnection(hostZone.ChainId, clientId, hostZone.ConnectionId)
	}

	// Return callback to check store after upgrade
	return func() {
		// Check that the module and ICA accounts were registered
		s.checkCommunityPoolModuleAccountsRegistered(v17.GaiaChainId)
		s.checkCommunityPoolModuleAccountsRegistered(v17.OsmosisChainId)

		s.checkCommunityPoolICAAccountsRegistered(v17.GaiaChainId)
		s.checkCommunityPoolICAAccountsRegistered(v17.OsmosisChainId)

		// Check that the redemption rate bounds were set
		gaiaHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v17.GaiaChainId)
		s.Require().True(found)

		s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("1.045"), gaiaHostZone.MinRedemptionRate, "gaia min outer") // 1.1 - 5% = 1.045
		s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("1.210"), gaiaHostZone.MaxRedemptionRate, "gaia max outer") // 1.1 + 10% = 1.21

		osmoHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "osmosis-1")
		s.Require().True(found)

		s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("1.140"), osmoHostZone.MinRedemptionRate, "osmo min outer") // 1.2 - 5% = 1.140
		s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("1.344"), osmoHostZone.MaxRedemptionRate, "osmo max outer") // 1.2 + 12% = 1.344

		// Check that there are no slash queries in progress
		for _, hostZone := range s.App.StakeibcKeeper.GetAllHostZone(s.Ctx) {
			for _, validator := range hostZone.Validators {
				s.Require().False(validator.SlashQueryInProgress, "slash query in progress should have been set to false")
			}
		}
	}
}

func (s *UpgradeTestSuite) SetupMigrateUnbondingRecords() func() {
	// Create EURs for two host zones
	// 2 HZU on each host zone will trigger URR updates
	//   - UNBONDING_QUEUE
	//   - UNBONDING_IN_PROGRESS
	//   - EXIT_TRANSFER_QUEUE
	//   - EXIT_TRANSFER_IN_PROGRESS
	// 2 HZU on each host zone will not trigger URR updates
	//   - 1 HZU has 0 NativeTokenAmount
	//   - 1 HZU has status CLAIMABLE

	nativeTokenAmount := int64(2000000)
	stTokenAmount := int64(1000000)
	URRAmount := int64(500)

	// create mockURRIds 1 through 6 and store the URRs
	for i := 1; i <= 6; i++ {
		mockURRId := strconv.Itoa(i)
		mockURR := recordtypes.UserRedemptionRecord{
			Id:                mockURRId,
			NativeTokenAmount: sdkmath.NewInt(URRAmount),
		}
		s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, mockURR)
	}

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			// HZUs that will trigger URR updates
			{
				HostZoneId:            v17.GaiaChainId,
				NativeTokenAmount:     sdkmath.NewInt(nativeTokenAmount),
				StTokenAmount:         sdkmath.NewInt(stTokenAmount),
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				UserRedemptionRecords: []string{"1"},
			},
			{
				HostZoneId:            SommelierChainId,
				NativeTokenAmount:     sdkmath.NewInt(nativeTokenAmount),
				StTokenAmount:         sdkmath.NewInt(stTokenAmount),
				Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				UserRedemptionRecords: []string{"2"},
			},
		},
	}
	epochUnbondingRecord2 := recordtypes.EpochUnbondingRecord{
		EpochNumber: 2,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			// HZUs that will trigger URR updates
			{
				HostZoneId:            v17.GaiaChainId,
				NativeTokenAmount:     sdkmath.NewInt(nativeTokenAmount),
				StTokenAmount:         sdkmath.NewInt(stTokenAmount),
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				UserRedemptionRecords: []string{"3"},
			},
			{
				HostZoneId:            SommelierChainId,
				NativeTokenAmount:     sdkmath.NewInt(nativeTokenAmount),
				StTokenAmount:         sdkmath.NewInt(stTokenAmount),
				Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS,
				UserRedemptionRecords: []string{"4"},
			},
		},
	}
	epochUnbondingRecord3 := recordtypes.EpochUnbondingRecord{
		EpochNumber: 3,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			// HZUs that will not trigger URR updates
			{
				HostZoneId: v17.GaiaChainId,
				// Will not trigger update because NativeTokenAmount is 0
				NativeTokenAmount:     sdkmath.NewInt(0),
				StTokenAmount:         sdkmath.NewInt(stTokenAmount),
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				UserRedemptionRecords: []string{"5"},
			},
			{
				HostZoneId:        v17.GaiaChainId,
				NativeTokenAmount: sdkmath.NewInt(nativeTokenAmount),
				StTokenAmount:     sdkmath.NewInt(stTokenAmount),
				// Will not trigger update because status is CLAIMABLE
				Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				UserRedemptionRecords: []string{"6"},
			},
		},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord2)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord3)

	return func() {
		// conversionRate is stTokenAmount / nativeTokenAmount
		conversionRate := sdkmath.LegacyNewDec(stTokenAmount).Quo(sdkmath.LegacyNewDec(nativeTokenAmount))
		expectedConversionRate := sdkmath.LegacyMustNewDecFromStr("0.5")
		s.Require().Equal(expectedConversionRate, conversionRate, "expected conversion rate (1/redemption rate)")

		// stTokenAmount is conversionRate * URRAmount
		stTokenAmount := conversionRate.Mul(sdkmath.LegacyNewDec(URRAmount)).RoundInt()
		expectedStTokenAmount := sdkmath.NewInt(250)
		s.Require().Equal(stTokenAmount, expectedStTokenAmount, "expected st token amount")

		// Verify URR stToken amounts are set correctly for records 1 through 4
		for i := 1; i <= 4; i++ {
			mockURRId := strconv.Itoa(i)
			mockURR, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, mockURRId)
			s.Require().True(found)
			s.Require().Equal(expectedStTokenAmount, mockURR.StTokenAmount, "URR %s - st token amount", mockURRId)
		}

		// Verify URRs with status CLAIMABLE are skipped (record 5)
		// Verify HZUs with 0 NativeTokenAmount are skipped (record 6)
		for i := 5; i <= 6; i++ {
			mockURRId := strconv.Itoa(i)
			mockURR, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, mockURRId)
			s.Require().True(found)
			// verify the amount was not updated
			s.Require().Equal(sdkmath.NewInt(0), mockURR.StTokenAmount, "URR %s - st token amount", mockURRId)
		}
	}
}

func (s *UpgradeTestSuite) SetupRateLimitsBeforeUpgrade() func() {
	gaiaChannelId := "channel-0"

	initialThreshold := sdkmath.OneInt()
	initialFlow := sdkmath.NewInt(10)
	initialChannelValue := sdkmath.NewInt(100)
	updatedChannelValue := sdkmath.NewInt(200)

	rateLimits := []AddRateLimits{
		{
			// Gaia rate limit
			// Threshold should be updated, new gaia RL added
			Denom:     StAtom,
			ChannelId: gaiaChannelId,
		},
		{
			// Osmo rate limit
			// Threshold should be updated, no new RL added
			Denom:     StOsmo,
			ChannelId: v17.OsmosisTransferChannelId,
		},
		{
			// Somm rate limit
			// Should be removed
			Denom:     StSomm,
			ChannelId: "channel-10",
		},
	}

	// Add rate limits
	// No need to register host zones since they're initialized in the setup host zone function
	for _, rateLimit := range rateLimits {
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     rateLimit.Denom,
				ChannelId: rateLimit.ChannelId,
			},
			Quota: &ratelimittypes.Quota{
				MaxPercentSend: initialThreshold,
				MaxPercentRecv: initialThreshold,
				DurationHours:  10,
			},
			Flow: &ratelimittypes.Flow{
				Outflow:      initialFlow,
				Inflow:       initialFlow,
				ChannelValue: initialChannelValue,
			},
		})

		// mint the token for the channel value
		s.FundAccount(s.TestAccs[0], sdk.NewCoin(rateLimit.Denom, updatedChannelValue))
	}

	// Return callback to check store after upgrade
	return func() {
		// Check that we have 3 RLs
		//   (1) stuosmo on Stride -> Osmosis
		//   (2) stuatom on Stride -> Gaia
		//   (3) stuatom on Stride -> Osmosis
		acutalRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
		s.Require().Len(acutalRateLimits, 3, "there should be 3 rate limits at the end")

		// Check the stosmo rate limit
		stOsmoRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, StOsmo, v17.OsmosisTransferChannelId)
		s.Require().True(found)

		osmoThreshold := v17.UpdatedRateLimits[v17.OsmosisChainId]
		s.Require().Equal(osmoThreshold, stOsmoRateLimit.Quota.MaxPercentSend, "stosmo max percent send")
		s.Require().Equal(osmoThreshold, stOsmoRateLimit.Quota.MaxPercentRecv, "stosmo max percent recv")
		s.Require().Equal(initialFlow, stOsmoRateLimit.Flow.Outflow, "stosmo outflow")
		s.Require().Equal(initialFlow, stOsmoRateLimit.Flow.Inflow, "stosmo inflow")
		s.Require().Equal(initialChannelValue, stOsmoRateLimit.Flow.ChannelValue, "stosmo channel value")

		// Check the stuatom -> Gaia rate limit
		stAtomToGaiaRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, StAtom, gaiaChannelId)
		s.Require().True(found)

		atomThreshold := v17.UpdatedRateLimits[v17.GaiaChainId]
		s.Require().Equal(atomThreshold, stAtomToGaiaRateLimit.Quota.MaxPercentSend, "statom -> gaia max percent send")
		s.Require().Equal(atomThreshold, stAtomToGaiaRateLimit.Quota.MaxPercentRecv, "statom -> gaia max percent recv")
		s.Require().Equal(initialFlow, stAtomToGaiaRateLimit.Flow.Outflow, "statom -> gaia outflow")
		s.Require().Equal(initialFlow, stAtomToGaiaRateLimit.Flow.Inflow, "statom -> gaia inflow")
		s.Require().Equal(initialChannelValue, stAtomToGaiaRateLimit.Flow.ChannelValue, "statom -> gaia channel value")

		// Check the stuatom -> Osmo rate limit
		// The flow should be reset to 0 and the channel value should update
		stAtomToOsmoRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, StAtom, v17.OsmosisTransferChannelId)
		s.Require().True(found)

		s.Require().Equal(atomThreshold, stAtomToOsmoRateLimit.Quota.MaxPercentSend, "statom -> osmo max percent send")
		s.Require().Equal(atomThreshold, stAtomToOsmoRateLimit.Quota.MaxPercentRecv, "statom -> osmo max percent recv")

		s.Require().Zero(stAtomToOsmoRateLimit.Flow.Outflow.Int64(), "statom -> osmo outflow")
		s.Require().Zero(stAtomToOsmoRateLimit.Flow.Inflow.Int64(), "statom -> osmo inflow")
		s.Require().Equal(updatedChannelValue, stAtomToOsmoRateLimit.Flow.ChannelValue, "statom -> osmo channel value")
	}
}

func (s *UpgradeTestSuite) SetupCommunityPoolTaxBeforeUpgrade() func() {
	// Set initial community pool tax to 2%
	initialTax := sdkmath.LegacyMustNewDecFromStr("0.02")
	params, err := s.App.DistrKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err, "no error expected when getting params")
	params.CommunityTax = initialTax
	err = s.App.DistrKeeper.Params.Set(s.Ctx, params)
	s.Require().NoError(err, "no error expected when setting params")

	// Return callback to check store after upgrade
	return func() {
		// Confirm the tax increased
		updatedParams, err := s.App.DistrKeeper.Params.Get(s.Ctx)
		s.Require().NoError(err, "no error expected when getting params")
		s.Require().Equal(v17.CommunityPoolTax.String(), updatedParams.CommunityTax.String(),
			"community pool tax should have been updated")
	}
}

func (s *UpgradeTestSuite) SetupQueriesBeforeUpgrade() func() {
	// Set two queries - one for a slash, and one that's not for a slash
	queries := []icqtypes.Query{
		{
			Id:         "query-1",
			CallbackId: keeper.ICQCallbackID_Delegation,
		},
		{
			Id:         "query-2",
			CallbackId: keeper.ICQCallbackID_Calibrate,
		},
	}
	for _, query := range queries {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}

	// Return callback to check store after upgrade
	return func() {
		// Confirm one query was removed
		remainingQueries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
		s.Require().Len(remainingQueries, 1, "there should be only 1 remaining query")

		// Confirm the slash query was removed
		_, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, "query-1")
		s.Require().False(found, "slash query should have been removed")

		// Confirm the other query is still there
		_, found = s.App.InterchainqueryKeeper.GetQuery(s.Ctx, "query-2")
		s.Require().True(found, "non-slash query should not have been removed")
	}
}

func (s *UpgradeTestSuite) SetupProp225BeforeUpgrade() func() {
	// Grab the community pool growth address and balance
	communityPoolGrowthAddress := sdk.MustAccAddressFromBech32(v17.CommunityPoolGrowthAddress)
	// Set the balance to 3x the amount that will be transferred
	newCoin := sdk.NewCoin(v17.Ustrd, v17.Prop225TransferAmount.MulRaw(3))
	s.FundAccount(communityPoolGrowthAddress, newCoin)
	originalCommunityGrowthBalance := s.App.BankKeeper.GetBalance(s.Ctx, communityPoolGrowthAddress, v17.Ustrd)

	// Grab the liquidity receiver address and balance
	liquidityReceiverAddress := sdk.MustAccAddressFromBech32(v17.LiquidityReceiver)
	originalLiquidityReceiverBalance := s.App.BankKeeper.GetBalance(s.Ctx, liquidityReceiverAddress, v17.Ustrd)

	// grab how much we want to transfer
	transferAmount := v17.Prop225TransferAmount.Int64()

	// Return callback to check store after upgrade
	return func() {
		// verify funds left community growth
		newCommunityGrowthBalance := s.App.BankKeeper.GetBalance(s.Ctx, communityPoolGrowthAddress, v17.Ustrd)
		communityGrowthBalanceChange := originalCommunityGrowthBalance.Sub(newCommunityGrowthBalance)
		s.Require().Equal(transferAmount, communityGrowthBalanceChange.Amount.Int64(), "community growth decreased by correct amount")

		// verify funds entered liquidity custodian
		newLiquidityCustodianBalance := s.App.BankKeeper.GetBalance(s.Ctx, liquidityReceiverAddress, v17.Ustrd)
		liquidityCustodianBalanceChange := newLiquidityCustodianBalance.Sub(originalLiquidityReceiverBalance).Amount.Int64()
		s.Require().Equal(transferAmount, liquidityCustodianBalanceChange, "custodian balance increased by correct amount")
	}
}

func (s *UpgradeTestSuite) TestRegisterCommunityPoolAddresses() {
	// Create 3 host zones, with empty ICA addresses
	chainIds := []string{}
	for i := 1; i <= 3; i++ {
		chainId := fmt.Sprintf("chain-%d", i)
		connectionId := fmt.Sprintf("connection-%d", i)
		clientId := fmt.Sprintf("07-tendermint-%d", i)

		s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
			ChainId:      chainId,
			ConnectionId: connectionId,
		})
		chainIds = append(chainIds, chainId)

		// Mock out the consensus state to test registering an interchain account
		s.MockClientAndConnection(chainId, clientId, connectionId)
	}

	// Register the accounts
	err := v17.RegisterCommunityPoolAddresses(s.Ctx, s.App.StakeibcKeeper)
	s.Require().NoError(err, "no error expected when registering ICA addresses")

	// Confirm the module accounts and ICAs were registered
	for _, chainId := range chainIds {
		s.checkCommunityPoolModuleAccountsRegistered(chainId)
		s.checkCommunityPoolICAAccountsRegistered(chainId)
	}
}

func (s *UpgradeTestSuite) TestDeleteAllStaleQueries() {
	// Create queries - half of which are for slashed (CallbackId: Delegation)
	initialQueries := 10
	for i := 1; i <= initialQueries; i++ {
		queryId := fmt.Sprintf("query-%d", i)

		// Alternate slash queries vs other query
		callbackId := stakeibckeeper.ICQCallbackID_Delegation
		if i%2 == 0 {
			callbackId = stakeibckeeper.ICQCallbackID_FeeBalance // arbitrary
		}

		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqtypes.Query{
			Id:         queryId,
			CallbackId: callbackId,
		})
	}

	// Delete stale queries
	v17.DeleteAllStaleQueries(s.Ctx, s.App.InterchainqueryKeeper)

	// Check that only half the queries are remaining and none are for slashes
	remainingQueries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Equal(initialQueries/2, len(remainingQueries), "half the queries should have been removed")

	for _, query := range remainingQueries {
		s.Require().NotEqual(stakeibckeeper.ICQCallbackID_Delegation, query.CallbackId,
			"all slash queries should have been removed")
	}
}

func (s *UpgradeTestSuite) TestResetSlashQueryInProgress() {
	// Set multiple host zones, each with multiple validators that have slash query in progress set to true
	for hostIndex := 1; hostIndex <= 3; hostIndex++ {
		chainId := fmt.Sprintf("chain-%d", hostIndex)

		validators := []*stakeibctypes.Validator{}
		for validatorIndex := 1; validatorIndex <= 6; validatorIndex++ {
			address := fmt.Sprintf("val-%d", validatorIndex)

			inProgress := true
			if validatorIndex%2 == 0 {
				inProgress = false
			}

			validators = append(validators, &stakeibctypes.Validator{
				Address:              address,
				SlashQueryInProgress: inProgress,
			})
		}

		s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
			ChainId:    chainId,
			Validators: validators,
		})
	}

	// Reset the slash queries
	v17.ResetSlashQueryInProgress(s.Ctx, s.App.StakeibcKeeper)

	// Confirm they were all reset
	for _, hostZone := range s.App.StakeibcKeeper.GetAllHostZone(s.Ctx) {
		for _, validator := range hostZone.Validators {
			s.Require().False(validator.SlashQueryInProgress,
				"%s %s - slash query in progress should have been reset", hostZone.ChainId, validator.Address)
		}
	}
}

func (s *UpgradeTestSuite) TestExecuteProp223() {
	// Set initial community pool tax to 2%
	initialTax := sdkmath.LegacyMustNewDecFromStr("0.02")
	params, err := s.App.DistrKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err, "no error expected when getting params")
	params.CommunityTax = initialTax
	err = s.App.DistrKeeper.Params.Set(s.Ctx, params)
	s.Require().NoError(err, "no error expected when setting params")

	// Increase the tax
	err = v17.ExecuteProp223(s.Ctx, s.App.DistrKeeper)
	s.Require().NoError(err, "no error expected when increasing community pool tax")

	// Confirm it increased
	updatedParams, err := s.App.DistrKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err, "no error expected when getting params")
	s.Require().Equal(v17.CommunityPoolTax.String(), updatedParams.CommunityTax.String(),
		"community pool tax should have been updated")
}

func (s *UpgradeTestSuite) TestUpdateRedemptionRateBounds() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			// Max outer for osmo uses 12% instead of 10%
			ChainId:                        v17.OsmosisChainId,
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.4000"), // 1.25 + 12% = 1.400
		},
	}

	// Create a host zone for each test case
	for _, tc := range testCases {
		hostZone := stakeibctypes.HostZone{
			ChainId:        tc.ChainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Update the redemption rate bounds
	v17.UpdateRedemptionRateBounds(s.Ctx, s.App.StakeibcKeeper)

	// Confirm they were all updated
	for _, tc := range testCases {
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.ChainId)
		s.Require().True(found)

		s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "%s - min outer", tc.ChainId)
		s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "%s - max outer", tc.ChainId)
	}
}

func (s *UpgradeTestSuite) TestUpdateRateLimitThresholds() {
	initialThreshold := sdkmath.OneInt()

	// Define test cases consisting of an initial redemption rates and expected bounds
	testCases := map[string]UpdateRateLimits{
		"cosmoshub": {
			// 15% threshold
			ChainId:        "cosmoshub-4",
			ChannelId:      "channel-0",
			HostDenom:      "uatom",
			RateLimitDenom: "stuatom",
			Duration:       10,
			Threshold:      sdkmath.NewInt(15),
		},
		"osmosis": {
			// 15% threshold
			ChainId:        "osmosis-1",
			ChannelId:      "channel-1",
			HostDenom:      "uosmo",
			RateLimitDenom: "stuosmo",
			Duration:       20,
			Threshold:      sdkmath.NewInt(15),
		},
		"juno": {
			// No denom on matching host
			ChainId:        "juno-1",
			ChannelId:      "channel-2",
			HostDenom:      "ujuno",
			RateLimitDenom: "different-denom",
			Duration:       30,
		},
		"sommelier": {
			// Rate limit should get removed
			ChainId:        "sommelier-3",
			ChannelId:      "channel-3",
			HostDenom:      "usomm",
			RateLimitDenom: "stusomm",
			Duration:       40,
		},
	}

	// Set rate limits and host zones
	for _, tc := range testCases {
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     tc.RateLimitDenom,
				ChannelId: tc.ChannelId,
			},
			Quota: &ratelimittypes.Quota{
				MaxPercentSend: initialThreshold,
				MaxPercentRecv: initialThreshold,
				DurationHours:  tc.Duration,
			},
		})

		s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
			ChainId:   tc.ChainId,
			HostDenom: tc.HostDenom,
		})
	}

	// Update rate limits
	v17.UpdateRateLimitThresholds(s.Ctx, s.App.StakeibcKeeper, s.App.RatelimitKeeper)

	// Check that the osmo and gaia rate limits were updated
	for _, chainName := range []string{"cosmoshub", "osmosis"} {
		testCase := testCases[chainName]
		actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, testCase.RateLimitDenom, testCase.ChannelId)
		s.Require().True(found, "rate limit should have been found")

		// Check that the thresholds were updated
		s.Require().Equal(testCase.Threshold, actualRateLimit.Quota.MaxPercentSend, "%s - max percent send", chainName)
		s.Require().Equal(testCase.Threshold, actualRateLimit.Quota.MaxPercentRecv, "%s - max percent recv", chainName)
		s.Require().Equal(testCase.Duration, actualRateLimit.Quota.DurationHours, "%s - duration", chainName)
	}

	// Check that the juno rate limit was not touched
	// (since there was no matching host denom
	junoTestCase := testCases["juno"]
	actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, junoTestCase.RateLimitDenom, junoTestCase.ChannelId)
	s.Require().True(found, "juno rate limit should have been found")

	s.Require().Equal(initialThreshold, actualRateLimit.Quota.MaxPercentSend, "juno max percent send")
	s.Require().Equal(initialThreshold, actualRateLimit.Quota.MaxPercentRecv, "juno max percent recv")

	// Check that the somm rate limit was removed
	sommTestCase := testCases["sommelier"]
	_, found = s.App.RatelimitKeeper.GetRateLimit(s.Ctx, sommTestCase.RateLimitDenom, sommTestCase.ChannelId)
	s.Require().False(found, "somm rate limit should have been removed")
}

func (s *UpgradeTestSuite) TestAddRateLimitToOsmosis() {
	initialThreshold := sdkmath.OneInt()
	initialFlow := sdkmath.NewInt(100)
	initialDuration := uint64(24)
	initialChannelValue := sdkmath.NewInt(1000)

	// Define the test cases for adding new rate limits
	testCases := map[string]AddRateLimits{
		"cosmoshub": {
			// Will add a new rate limit to osmo
			Denom:        "stuatom",
			ChannelId:    "channel-0",
			ChannelValue: sdkmath.NewInt(100),
		},
		"osmosis": {
			// Rate limit already exists, should not be touched
			Denom:        "stuosmo",
			ChannelId:    v17.OsmosisTransferChannelId,
			ChannelValue: sdkmath.NewInt(300),
		},
		"stargaze": {
			// Will add a new rate limit to stars
			Denom:        "stustars",
			ChannelId:    "channel-1",
			ChannelValue: sdkmath.NewInt(200),
		},
	}

	// Setup the initial rate limits
	for _, tc := range testCases {
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     tc.Denom,
				ChannelId: tc.ChannelId,
			},
			Quota: &ratelimittypes.Quota{
				MaxPercentSend: initialThreshold,
				MaxPercentRecv: initialThreshold,
				DurationHours:  initialDuration,
			},
			Flow: &ratelimittypes.Flow{
				Outflow:      initialFlow,
				Inflow:       initialFlow,
				ChannelValue: initialChannelValue,
			},
		})

		// mint tokens so there's a supply for the channel value
		s.FundAccount(s.TestAccs[0], sdk.NewCoin(tc.Denom, tc.ChannelValue))
	}

	// Add the rate limits to osmo
	err := v17.AddRateLimitToOsmosis(s.Ctx, s.App.RatelimitKeeper)
	s.Require().NoError(err, "no error expected when adding rate limit to osmosis")

	// Check that we have new rate limits for gaia and stars
	newRateLimits := []string{"cosmoshub", "stargaze"}
	updatedRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Equal(len(testCases)+len(newRateLimits), len(updatedRateLimits), "number of ending rate limits")

	for _, chainName := range newRateLimits {
		testCase := testCases[chainName]

		// Confirm the new rate limit was created
		actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, testCase.Denom, v17.OsmosisTransferChannelId)
		s.Require().True(found, "new rate limit for %s to osmosis should have been found", chainName)

		// Confirm the thresholds remained the same
		s.Require().Equal(initialThreshold, actualRateLimit.Quota.MaxPercentSend, "%s - max percent send", chainName)
		s.Require().Equal(initialThreshold, actualRateLimit.Quota.MaxPercentRecv, "%s - max percent recv", chainName)
		s.Require().Equal(initialDuration, actualRateLimit.Quota.DurationHours, "%s - duration", chainName)

		// Confirm the flow as reset
		s.Require().Zero(actualRateLimit.Flow.Outflow.Int64(), "%s - outflow", chainName)
		s.Require().Zero(actualRateLimit.Flow.Inflow.Int64(), "%s - inflow", chainName)
		s.Require().Equal(testCase.ChannelValue, actualRateLimit.Flow.ChannelValue, "%s - channel value", chainName)
	}

	// Confirm the osmo rate limit was not touched
	osmoTestCase := testCases["osmosis"]
	osmoRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, osmoTestCase.Denom, v17.OsmosisTransferChannelId)
	s.Require().True(found, "rate limit for osmosis should have been found")

	s.Require().Equal(osmoRateLimit.Flow.Outflow, osmoRateLimit.Flow.Outflow, "osmos outflow")
	s.Require().Equal(osmoRateLimit.Flow.Inflow, osmoRateLimit.Flow.Inflow, "osmos inflow")
	s.Require().Equal(osmoRateLimit.Flow.ChannelValue, osmoRateLimit.Flow.ChannelValue, "osmos channel value")

	// Add a rate limit with zero channel value and confirm we cannot add a rate limit with that denom to osmosis
	nonExistentDenom := "denom"
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, ratelimittypes.RateLimit{
		Path: &ratelimittypes.Path{
			Denom:     nonExistentDenom,
			ChannelId: "channel-6",
		},
	})

	err = v17.AddRateLimitToOsmosis(s.Ctx, s.App.RatelimitKeeper)
	s.Require().ErrorContains(err, "channel value is zero")
}
