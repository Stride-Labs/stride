package v10_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	v10 "github.com/Stride-Labs/stride/v14/app/upgrades/v10"
	"github.com/Stride-Labs/stride/v14/utils"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	ratelimittypes "github.com/Stride-Labs/stride/v14/x/ratelimit/types"
	recordskeeper "github.com/Stride-Labs/stride/v14/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	deprecatedproto "github.com/golang/protobuf/proto" //nolint:staticcheck

	claimtypes "github.com/Stride-Labs/stride/v14/x/claim/types"
)

var initialRateLimitChannelValue = sdk.NewInt(1_000_000)

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
	dummyUpgradeHeight := int64(5)

	// Remove localhost client from client keeper
	clientParams := s.App.IBCKeeper.ClientKeeper.GetParams(s.Ctx)
	clientParams.AllowedClients = []string{}
	s.App.IBCKeeper.ClientKeeper.SetParams(s.Ctx, clientParams)

	// Fund the community pool growth address
	communityGrowthAddress := sdk.MustAccAddressFromBech32(v10.CommunityPoolGrowthAddress)
	s.FundAccount(communityGrowthAddress, sdk.NewCoin(v10.Ustrd, v10.BadKidsTransferAmount))

	// Add cosmoshub to check that a rate limit is added
	gaiaChainId := "cosmoshub-4"
	atom := "uatom"
	stAtom := "st" + atom
	transferChannelId := "channel-0"
	s.setupRateLimitedHostZone(gaiaChainId, stAtom, transferChannelId)

	// Create reward collector account for rate limit whitelist
	rewardCollectorAddress := s.createRewardCollectorModuleAccount()

	// Submit upgrade
	s.ConfirmUpgradeSucceededs("v10", dummyUpgradeHeight)

	// Check mint parameters after upgrade
	proportions := s.App.MintKeeper.GetParams(s.Ctx).DistributionProportions

	s.Require().Equal(v10.StakingProportion,
		proportions.Staking.String()[:9], "staking")

	s.Require().Equal(v10.CommunityPoolGrowthProportion,
		proportions.CommunityPoolGrowth.String()[:9], "community pool growth")

	s.Require().Equal(v10.StrategicReserveProportion,
		proportions.StrategicReserve.String()[:9], "strategic reserve")

	s.Require().Equal(v10.CommunityPoolSecurityBudgetProportion,
		proportions.CommunityPoolSecurityBudget.String()[:9], "community pool security")

	// Check initial deposit ratio
	govParams := s.App.GovKeeper.GetParams(s.Ctx)
	s.Require().Equal(v10.MinInitialDepositRatio, govParams.MinInitialDepositRatio, "min initial deposit ratio")

	// Check localhost client was added
	clientParams = s.App.IBCKeeper.ClientKeeper.GetParams(s.Ctx)
	s.Require().Contains(clientParams.AllowedClients, "09-localhost")

	// Check the transfer was successful
	communityGrowthBalance := s.App.BankKeeper.GetBalance(s.Ctx, communityGrowthAddress, v10.Ustrd)
	s.Require().Zero(communityGrowthBalance.Amount.Int64(), "community growth balance after transfer")

	badKidsCustodianAddress := sdk.MustAccAddressFromBech32(v10.BadKidsCustodian)
	badKidsCustodianBalance := s.App.BankKeeper.GetBalance(s.Ctx, badKidsCustodianAddress, v10.Ustrd)
	s.Require().Equal(v10.BadKidsTransferAmount.Int64(), badKidsCustodianBalance.Amount.Int64(), "bad kids balance")

	// Check that the rate limit was added
	rateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Len(rateLimits, 1, "one rate limit should have been added")
	s.validateRateLimit(rateLimits[0], gaiaChainId, stAtom, transferChannelId)
	s.validateWhitelists(gaiaChainId, rewardCollectorAddress)
}

func (s *UpgradeTestSuite) createCallbackData(id string, callback deprecatedproto.Message) icacallbackstypes.CallbackData {
	return icacallbackstypes.CallbackData{
		CallbackId:   id,
		CallbackArgs: s.mustMarshalCallback(callback),
	}
}

func (s *UpgradeTestSuite) mustMarshalCallback(callback deprecatedproto.Message) []byte {
	callbackBz, err := deprecatedproto.Marshal(callback)
	s.Require().NoError(err)
	return callbackBz
}

func (s *UpgradeTestSuite) mustUnmarshalCallback(callbackBz []byte, callback cosmosproto.Message) {
	err := cosmosproto.Unmarshal(callbackBz, callback)
	s.Require().NoError(err)
}

func (s *UpgradeTestSuite) TestMigrateCallbackData() {
	// Build dummy callback data for each callback type
	initialClaimCallbackArgs := stakeibctypes.ClaimCallback{
		UserRedemptionRecordId: "record-0",
		ChainId:                "chain-0",
		EpochNumber:            1,
	}
	initialDelegateCallbackArgs := stakeibctypes.DelegateCallback{
		HostZoneId:      "host-0",
		DepositRecordId: 1,
		SplitDelegations: []*types.SplitDelegation{{
			Validator: "val-0",
			Amount:    sdkmath.NewInt(1),
		}},
	}
	initialRebalanceCallbackArgs := stakeibctypes.RebalanceCallback{
		HostZoneId: "host-0",
		Rebalancings: []*stakeibctypes.Rebalancing{
			{
				SrcValidator: "val-0",
				DstValidator: "val-1",
				Amt:          sdkmath.NewInt(1),
			},
		},
	}
	initialRedemptionCallbackArgs := stakeibctypes.RedemptionCallback{
		HostZoneId:              "host-0",
		EpochUnbondingRecordIds: []uint64{1, 2, 3},
	}
	initialReinvestCallbackArgs := stakeibctypes.ReinvestCallback{
		HostZoneId:     "host-0",
		ReinvestAmount: sdk.NewCoin("denom", sdkmath.NewInt(1)),
	}
	initialUndelegateCallbackArgs := stakeibctypes.UndelegateCallback{
		HostZoneId: "host-0",
		SplitDelegations: []*types.SplitDelegation{{
			Validator: "val-0",
			Amount:    sdkmath.NewInt(1),
		}},
	}
	initialTransferCallbackArgs := recordstypes.TransferCallback{
		DepositRecordId: 1,
	}

	// Store the callback data
	initialCallbackData := []icacallbackstypes.CallbackData{
		s.createCallbackData(stakeibckeeper.ICACallbackID_Claim, &initialClaimCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Delegate, &initialDelegateCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Rebalance, &initialRebalanceCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Redemption, &initialRedemptionCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Reinvest, &initialReinvestCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Undelegate, &initialUndelegateCallbackArgs),
		s.createCallbackData(recordskeeper.IBCCallbacksID_NativeTransfer, &initialTransferCallbackArgs),
	}
	for i := range initialCallbackData {
		initialCallbackData[i].CallbackKey = fmt.Sprintf("key-%d", i)
		initialCallbackData[i].PortId = fmt.Sprintf("port-%d", i)
		initialCallbackData[i].ChannelId = fmt.Sprintf("channel-%d", i)
		s.App.IcacallbacksKeeper.SetCallbackData(s.Ctx, initialCallbackData[i])
	}

	// Migrate the callbacks
	err := v10.MigrateCallbackData(s.Ctx, s.App.IcacallbacksKeeper)
	s.Require().NoError(err, "no error expected when migrating callback data")

	// Check that we can successfully unmarshal each callback with the new type
	finalCallbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(finalCallbackData, len(initialCallbackData), "callback data length")

	for i, finalCallback := range finalCallbackData {
		initialCallback := initialCallbackData[i]
		s.Require().Equal(initialCallback.CallbackId, finalCallback.CallbackId, "callback id for %d", i)

		callbackId := initialCallback.CallbackId
		s.Require().Equal(initialCallback.CallbackKey, finalCallback.CallbackKey, "callback key for %s", callbackId)
		s.Require().Equal(initialCallback.PortId, finalCallback.PortId, "callback port for %s", callbackId)
		s.Require().Equal(initialCallback.ChannelId, finalCallback.ChannelId, "callback channel for %s", callbackId)

		switch callbackId {
		case stakeibckeeper.ICACallbackID_Claim:
			var finalCallbackArgs stakeibctypes.ClaimCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialClaimCallbackArgs, finalCallbackArgs, "claim callback")

		case stakeibckeeper.ICACallbackID_Delegate:
			var finalCallbackArgs stakeibctypes.DelegateCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialDelegateCallbackArgs, finalCallbackArgs, "delegate callback")

		case stakeibckeeper.ICACallbackID_Rebalance:
			var finalCallbackArgs stakeibctypes.RebalanceCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialRebalanceCallbackArgs, finalCallbackArgs, "rebalance callback")

		case stakeibckeeper.ICACallbackID_Redemption:
			var finalCallbackArgs stakeibctypes.RedemptionCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialRedemptionCallbackArgs, finalCallbackArgs, "redemption callback")

		case stakeibckeeper.ICACallbackID_Reinvest:
			var finalCallbackArgs stakeibctypes.ReinvestCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialReinvestCallbackArgs, finalCallbackArgs, "reinvest callback")

		case stakeibckeeper.ICACallbackID_Undelegate:
			var finalCallbackArgs stakeibctypes.UndelegateCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialUndelegateCallbackArgs, finalCallbackArgs, "undelegate callback")

		case recordskeeper.IBCCallbacksID_NativeTransfer:
			var finalCallbackArgs recordstypes.TransferCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialTransferCallbackArgs, finalCallbackArgs, "transfer callback")
		}
	}
}

func (s *UpgradeTestSuite) TestMigrateDistributorAddress() {
	ck := s.App.ClaimKeeper

	// Airdrops
	strideAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "stride",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "stride-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(4143585840),
		DistributorAddress: "stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev",
	}

	gaiaAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "gaia",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "cosmoshub-4",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(191138794312),
		DistributorAddress: "stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz",
	}

	osmosisAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "osmosis",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "osmosis-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(72895369704),
		DistributorAddress: "stride1zlu2l3lx5tqvzspvjwsw9u0e907kelhqae3yhk",
	}

	junoAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "juno",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "juno-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(10967183382),
		DistributorAddress: "stride14k9g9zpgaycpey9840nnpa66l4nd6lu7g7t74c",
	}

	starsAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "stars",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "stargaze-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(1013798205),
		DistributorAddress: "stride12pum4adk5dhp32d90f8g8gfwujm0gwxqnrdlum",
	}

	evmosAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "evmos",
		AirdropStartTime:   time.Date(2023, 4, 3, 16, 0, 0, 0, time.UTC),
		AutopilotEnabled:   true,
		ChainId:            "evmos_9001-2",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(13491005333),
		DistributorAddress: "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j",
	}

	claimParams, err := ck.GetParams(s.Ctx)
	s.Require().NoError(err, "no error expected when getting claim params")

	// Set the airdrops on the claim params
	claimParams.Airdrops = []*claimtypes.Airdrop{&strideAirdrop, &gaiaAirdrop, &osmosisAirdrop, &junoAirdrop, &starsAirdrop, &evmosAirdrop}
	err = ck.SetParams(s.Ctx, claimParams)
	s.Require().NoError(err, "no error expected when setting claim params")

	// Migrate the airdrop distributor addresses
	err = v10.MigrateClaimDistributorAddress(s.Ctx, ck)
	s.Require().NoError(err, "no error expected when migrating distributor addresses")

	// Iterate allAirdrops and make sure the updated airdrops are equivalent, except for the distributor address
	for _, airdrop := range claimParams.Airdrops {
		airdropToCompare := ck.GetAirdropByIdentifier(s.Ctx, airdrop.AirdropIdentifier)
		s.Require().Equal(airdrop.AirdropIdentifier, airdropToCompare.AirdropIdentifier, "airdrop identifier")
		s.Require().Equal(airdrop.AirdropStartTime, airdropToCompare.AirdropStartTime, "airdrop start time")
		s.Require().Equal(airdrop.AutopilotEnabled, airdropToCompare.AutopilotEnabled, "airdrop autopilot enabled")
		s.Require().Equal(airdrop.ChainId, airdropToCompare.ChainId, "airdrop chain id")
		s.Require().Equal(airdrop.ClaimDenom, airdropToCompare.ClaimDenom, "airdrop claim denom")
		s.Require().Equal(airdrop.ClaimedSoFar, airdropToCompare.ClaimedSoFar, "airdrop claimed so far")

		newDistributorAddress := v10.NewDistributorAddresses[airdrop.AirdropIdentifier]
		s.Require().Equal(newDistributorAddress, airdropToCompare.DistributorAddress, "airdrop distributor address updated")
	}
}

func (s *UpgradeTestSuite) createRewardCollectorModuleAccount() string {
	rewardCollectorAddress := address.Module(stakeibctypes.RewardCollectorName, []byte(stakeibctypes.RewardCollectorName))
	err := utils.CreateModuleAccount(s.Ctx, s.App.AccountKeeper, rewardCollectorAddress)
	s.Require().NoError(err, "no error expected when creating reward collector module account")

	rewardCollector := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName)
	return rewardCollector.GetAddress().String()
}

func (s *UpgradeTestSuite) setupRateLimitedHostZone(chainId, stDenom, channelId string) {
	// Store host zone in stakeibc
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:              chainId,
		HostDenom:            strings.ReplaceAll(stDenom, "st", ""),
		TransferChannelId:    channelId,
		DepositAddress:       chainId + ".STAKEIBC",
		FeeIcaAddress:        chainId + ".FEE",
		DelegationIcaAddress: chainId + ".DELEGATION",
	})

	// Create the transfer channel
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, channelId, channeltypes.Channel{})

	// Mint the st denom so there's an initial channel value
	initialStAtom := sdk.NewCoin(stDenom, initialRateLimitChannelValue)
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(initialStAtom))
	s.Require().NoError(err)
}

func (s *UpgradeTestSuite) validateRateLimit(rateLimit ratelimittypes.RateLimit, chainId, denom, channelId string) {
	s.Require().Equal(denom, rateLimit.Path.Denom, "rate limit denom")
	s.Require().Equal(channelId, rateLimit.Path.ChannelId, "rate limit channel")
	description := fmt.Sprintf("%s - %s - %s", chainId, denom, channelId)

	expectedThreshold := v10.NewRateLimits[chainId].Int64()
	quota := rateLimit.Quota
	s.Require().Equal(expectedThreshold, quota.MaxPercentSend.Int64(), "%s - rate limit send threshold", description)
	s.Require().Equal(expectedThreshold, quota.MaxPercentRecv.Int64(), "%s - rate limit recv threshold", description)
	s.Require().Equal(v10.RateLimitDurationHours, quota.DurationHours, "%s - rate limit duration", description)

	flow := rateLimit.Flow
	s.Require().Zero(flow.Inflow.Int64(), "%s - rate limit inflow", description)
	s.Require().Zero(flow.Outflow.Int64(), "%s - rate limit outflow", description)
	s.Require().Equal(initialRateLimitChannelValue.Int64(), flow.ChannelValue.Int64(),
		"%s - rate limit channel value", description)
}

func (s *UpgradeTestSuite) validateWhitelists(chainId, rewardCollectorAddress string) {
	stakeibcModuleAccount := chainId + ".STAKEIBC"
	delegationAddress := chainId + ".DELEGATION"
	feeAddress := chainId + ".FEE"

	isWhitelisted := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, stakeibcModuleAccount, delegationAddress)
	s.Require().True(isWhitelisted, "%d - stakeibc module account -> delegation ICA whitelisted", chainId)

	isWhitelisted = s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, feeAddress, rewardCollectorAddress)
	s.Require().True(isWhitelisted, "%d - fee account -> reward collector", chainId)
}

func (s *UpgradeTestSuite) TestEnableRateLimits() {
	rewardCollectorAddress := s.createRewardCollectorModuleAccount()

	// Create a host zone for each new rate limited host
	rateLimitedHosts := utils.StringMapKeys(v10.NewRateLimits)
	for i, chainId := range rateLimitedHosts {
		denom := fmt.Sprintf("stdenom-%d", i)
		channelId := fmt.Sprintf("channel-%d", i)
		s.setupRateLimitedHostZone(chainId, denom, channelId)
	}

	// Add a host zone that's not in the rate limit map
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: "differnet-host-zone",
	})

	// Enable the rate limits
	err := v10.EnableRateLimits(s.Ctx, s.App.AccountKeeper, s.App.IBCKeeper.ChannelKeeper, s.App.RatelimitKeeper, s.App.StakeibcKeeper)
	s.Require().NoError(err, "no error expected when enabling new rate limits")

	// Confirm correct number of new rate limits
	allRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Equal(len(v10.NewRateLimits), len(allRateLimits), "number of new rate limits")

	// Verify each rate limit
	for i, chainId := range rateLimitedHosts {
		denom := fmt.Sprintf("stdenom-%d", i)
		channelId := fmt.Sprintf("channel-%d", i)

		rateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found, "rate limit for %s and %s should have been found", denom, channelId)
		s.validateRateLimit(rateLimit, chainId, denom, channelId)
	}

	// Verify each whitelist pair
	whitelistedPairs := s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx)
	s.Require().Equal(len(rateLimitedHosts)*2, len(whitelistedPairs), "two whitelisted address pairs per host")

	for _, chainId := range rateLimitedHosts {
		s.validateWhitelists(chainId, rewardCollectorAddress)
	}
}
