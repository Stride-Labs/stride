package v17_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	ratelimittypes "github.com/Stride-Labs/stride/v16/x/ratelimit/types"

	"github.com/Stride-Labs/stride/v16/app/apptesting"
	v17 "github.com/Stride-Labs/stride/v16/app/upgrades/v17"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinOuterRedemptionRate sdk.Dec
	ExpectedMinInnerRedemptionRate sdk.Dec
	ExpectedMaxInnerRedemptionRate sdk.Dec
	ExpectedMaxOuterRedemptionRate sdk.Dec
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

	// Confirm the module accounts were created and stored on each host
	for _, chainId := range chainIds {
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

	// Confirm the ICAs were registered
	// The addresses don't get set until the callback, but we can check the events and confirm they were registered
	for _, chainId := range chainIds {
		depositOwner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
		returnOwner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_COMMUNITY_POOL_RETURN)

		expectedDepositPortId, _ := icatypes.NewControllerPortID(depositOwner)
		expectedReturnPortId, _ := icatypes.NewControllerPortID(returnOwner)

		s.CheckEventValueEmitted(channeltypes.EventTypeChannelOpenInit, channeltypes.AttributeKeyPortID, expectedDepositPortId)
		s.CheckEventValueEmitted(channeltypes.EventTypeChannelOpenInit, channeltypes.AttributeKeyPortID, expectedReturnPortId)
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

func (s *UpgradeTestSuite) TestIncreaseCommunityPoolTax() {
	// Set initial community pool tax to 2%
	initialTax := sdk.MustNewDecFromStr("0.02")
	params := s.App.DistrKeeper.GetParams(s.Ctx)
	params.CommunityTax = initialTax
	err := s.App.DistrKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err, "no error expected when setting params")

	// Increase the tax
	err = v17.IncreaseCommunityPoolTax(s.Ctx, s.App.DistrKeeper)
	s.Require().NoError(err, "no error expected when increasing community pool tax")

	// Confirm it increased
	updatedParams := s.App.DistrKeeper.GetParams(s.Ctx)
	s.Require().Equal(v17.CommunityPoolTax.String(), updatedParams.CommunityTax.String(),
		"community pool tax should have been updated")
}

func (s *UpgradeTestSuite) TestUpdateRedemptionRateBounds() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("0.97"), // 1 - 3% = 0.97
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.05"), // 1 + 5% = 1.05
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.067"), // 1.1 - 3% = 1.067
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.155"), // 1.1 + 5% = 1.155
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			// Max outer for osmo uses 12% instead of 10%
			ChainId:                        v17.OsmosisChainId,
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.2125"), // 1.25 - 3% = 1.2125
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.3125"), // 1.25 + 5% = 1.3125
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.4000"), // 1.25 + 12% = 1.400
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
		s.Require().Equal(tc.ExpectedMinInnerRedemptionRate, hostZone.MinInnerRedemptionRate, "%s - min inner", tc.ChainId)
		s.Require().Equal(tc.ExpectedMaxInnerRedemptionRate, hostZone.MaxInnerRedemptionRate, "%s - max inner", tc.ChainId)
		s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "%s - max outer", tc.ChainId)
	}
}

func (s *UpgradeTestSuite) TestUpdateRateLimitThresholds() {
	initialThreshold := sdkmath.OneInt()

	// Define test cases consisting of an initial redemption rates and expected bounds
	testCases := map[string]UpdateRateLimits{
		"cosmoshub": {
			// 10% threshold
			ChainId:        "cosmoshub-4",
			ChannelId:      "channel-0",
			HostDenom:      "uatom",
			RateLimitDenom: "stuatom",
			Duration:       10,
			Threshold:      sdkmath.NewInt(10),
		},
		"osmosis": {
			// 10% threshold
			ChainId:        "osmosis-1",
			ChannelId:      "channel-1",
			HostDenom:      "uosmo",
			RateLimitDenom: "stuosmo",
			Duration:       20,
			Threshold:      sdkmath.NewInt(10),
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
	initialChannelValue := sdk.NewInt(1000)

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
