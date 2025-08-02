package v19_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"cosmossdk.io/store/prefix"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app"
	"github.com/Stride-Labs/stride/v28/app/apptesting"
	v19 "github.com/Stride-Labs/stride/v28/app/upgrades/v19"
	legacyratelimittypes "github.com/Stride-Labs/stride/v28/app/upgrades/v19/legacyratelimit/types"
)

var StTiaSupply = sdkmath.NewInt(1000)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// Setup state before upgrade
	checkMigratedRateLimits := s.SetupRateLimitMigration()
	checkStTiaRateLimits := s.SetupStTiaRateLimits()

	// Run through upgrade
	s.ConfirmUpgradeSucceeded(v19.UpgradeName)

	// Check state after upgrade
	checkMigratedRateLimits()
	checkStTiaRateLimits()
	s.CheckWasmPerms()
}

func (s *UpgradeTestSuite) SetupRateLimitMigration() func() {
	rateLimitStore := s.Ctx.KVStore(s.App.GetKey(ratelimittypes.StoreKey))
	cdc := app.MakeEncodingConfig().Codec

	denom := "denom"
	channelId := "channel-0"
	flow := sdkmath.NewInt(10)
	channelValue := sdkmath.NewInt(100)

	initialRateLimit := legacyratelimittypes.RateLimit{
		Path: &legacyratelimittypes.Path{
			Denom:     denom,
			ChannelId: channelId,
		},
		Flow: &legacyratelimittypes.Flow{
			Inflow:       flow,
			Outflow:      flow,
			ChannelValue: channelValue,
		},
		Quota: &legacyratelimittypes.Quota{
			MaxPercentSend: sdkmath.NewInt(10),
			MaxPercentRecv: sdkmath.NewInt(10),
			DurationHours:  24,
		},
	}

	expectedRateLimit := ratelimittypes.RateLimit{
		Path: &ratelimittypes.Path{
			Denom:     denom,
			ChannelId: channelId,
		},
		Flow: &ratelimittypes.Flow{
			Inflow:       flow,
			Outflow:      flow,
			ChannelValue: channelValue,
		},
		Quota: &ratelimittypes.Quota{
			MaxPercentSend: sdkmath.NewInt(10),
			MaxPercentRecv: sdkmath.NewInt(10),
			DurationHours:  24,
		},
	}

	initialRateLimitBz, err := cdc.Marshal(&initialRateLimit)
	s.Require().NoError(err)

	hostzoneStore := prefix.NewStore(rateLimitStore, ratelimittypes.RateLimitKeyPrefix)
	hostzoneStore.Set(ratelimittypes.GetRateLimitItemKey(denom, channelId), initialRateLimitBz)

	// Return a callback to check the state after the upgrade
	return func() {
		actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found, "rate limit should have been found")
		s.Require().Equal(expectedRateLimit, actualRateLimit, "rate limit after upgrade")
	}
}

func (s *UpgradeTestSuite) SetupStTiaRateLimits() func() {
	// mint sttia so that there is a channel value
	s.FundAccount(s.TestAccs[0], sdk.NewCoin(v19.StTiaDenom, StTiaSupply))

	// Mock out a channel for osmosis and celstia
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, v19.CelestiaTransferChannelId, channeltypes.Channel{})
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, v19.OsmosisTransferChannelId, channeltypes.Channel{})
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, v19.NeutronTransferChannelId, channeltypes.Channel{})

	// Return a callback to check the rate limits were added after the upgrade
	return func() {
		expectedRateLimitToCelestia := ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     v19.StTiaDenom,
				ChannelId: v19.CelestiaTransferChannelId,
			},
			Flow: &ratelimittypes.Flow{
				Inflow:       sdkmath.NewInt(0),
				Outflow:      sdkmath.NewInt(0),
				ChannelValue: StTiaSupply,
			},
			Quota: &ratelimittypes.Quota{
				MaxPercentSend: sdkmath.NewInt(10),
				MaxPercentRecv: sdkmath.NewInt(10),
				DurationHours:  24,
			},
		}

		expectedRateLimitToOsmosis := ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     v19.StTiaDenom,
				ChannelId: v19.OsmosisTransferChannelId,
			},
			Flow: &ratelimittypes.Flow{
				Inflow:       sdkmath.NewInt(0),
				Outflow:      sdkmath.NewInt(0),
				ChannelValue: StTiaSupply,
			},
			Quota: &ratelimittypes.Quota{
				MaxPercentSend: sdkmath.NewInt(10),
				MaxPercentRecv: sdkmath.NewInt(10),
				DurationHours:  24,
			},
		}

		expectedRateLimitToNeutron := ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     v19.StTiaDenom,
				ChannelId: v19.NeutronTransferChannelId,
			},
			Flow: &ratelimittypes.Flow{
				Inflow:       sdkmath.NewInt(0),
				Outflow:      sdkmath.NewInt(0),
				ChannelValue: StTiaSupply,
			},
			Quota: &ratelimittypes.Quota{
				MaxPercentSend: sdkmath.NewInt(10),
				MaxPercentRecv: sdkmath.NewInt(10),
				DurationHours:  24,
			},
		}

		actualCelestiaRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, v19.StTiaDenom, v19.CelestiaTransferChannelId)
		s.Require().True(found, "rate limit to celestia should have been found")
		s.Require().Equal(expectedRateLimitToCelestia, actualCelestiaRateLimit)

		actualOsmosisRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, v19.StTiaDenom, v19.OsmosisTransferChannelId)
		s.Require().True(found, "rate limit to osmosis should have been found")
		s.Require().Equal(expectedRateLimitToOsmosis, actualOsmosisRateLimit)

		actualRateLimitToNeutron, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, v19.StTiaDenom, v19.NeutronTransferChannelId)
		s.Require().True(found, "rate limit to osmosis should have been found")
		s.Require().Equal(expectedRateLimitToNeutron, actualRateLimitToNeutron)
	}
}

func (s *UpgradeTestSuite) CheckWasmPerms() {
	wasmParams := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, wasmParams.CodeUploadAccess.Permission, "upload permission")
	s.Require().Equal(v19.WasmAdmin, wasmParams.CodeUploadAccess.Addresses[0], "upload address")
	s.Require().Equal(wasmtypes.AccessTypeNobody, wasmParams.InstantiateDefaultPermission, "instantiate permission")
}
