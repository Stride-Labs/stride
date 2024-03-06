package v19_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v18/app"
	"github.com/Stride-Labs/stride/v18/app/apptesting"
	v19 "github.com/Stride-Labs/stride/v18/app/upgrades/v19"
	legacyratelimittypes "github.com/Stride-Labs/stride/v18/app/upgrades/v19/legacyratelimit/types"
)

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
	dummyUpgradeHeight := int64(5)

	// Setup state before upgrade
	checkRateLimitsAfterUpgrade := s.SetupRateLimitsBeforeUpgrade()

	// Run through upgrade
	s.ConfirmUpgradeSucceededs("v19", dummyUpgradeHeight)

	// Check state after upgrade
	checkRateLimitsAfterUpgrade()
	s.CheckWasmPerms()
}

func (s *UpgradeTestSuite) SetupRateLimitsBeforeUpgrade() func() {
	rateLimitStore := s.Ctx.KVStore(s.App.GetKey(ratelimittypes.StoreKey))
	cdc := app.MakeEncodingConfig().Marshaler

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

	return func() {
		actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found, "rate limit should have been found")
		s.Require().Equal(expectedRateLimit, actualRateLimit, "rate limit after upgrade")
	}
}

func (s *UpgradeTestSuite) CheckWasmPerms() {
	wasmParams := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, wasmParams.CodeUploadAccess.Permission, "upload permission")
	s.Require().Equal(v19.WasmAdmin, wasmParams.CodeUploadAccess.Addresses[0], "upload address")
	s.Require().Equal(wasmtypes.AccessTypeNobody, wasmParams.InstantiateDefaultPermission, "instantiate permission")
}
