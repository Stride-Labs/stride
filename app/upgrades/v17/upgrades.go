package v17

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"

	"github.com/Stride-Labs/stride/v16/utils"
	icqkeeper "github.com/Stride-Labs/stride/v16/x/interchainquery/keeper"
	ratelimitkeeper "github.com/Stride-Labs/stride/v16/x/ratelimit/keeper"
	ratelimittypes "github.com/Stride-Labs/stride/v16/x/ratelimit/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

var (
	UpgradeName = "v17"

	// Community pool tax updated from 2 -> 5%
	CommunityPoolTax = sdk.MustNewDecFromStr("0.05")

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateInnerMinAdjustment = sdk.MustNewDecFromStr("0.03")
	RedemptionRateInnerMaxAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdk.MustNewDecFromStr("0.02")

	// Rate limits updated according to TVL
	UpdatedRateLimits = map[string]sdkmath.Int{
		"comdex-1":     sdkmath.ZeroInt(),  // TVL: ~130k |     <1M  | No rate limit
		"cosmoshub-4":  sdkmath.NewInt(10), // TVL:  ~50M |    30M+  | 10% RL
		"evmos_9001-2": sdkmath.NewInt(50), // TVL:   ~3M | 1M-15M+  | 50% RL
		"injective-1":  sdkmath.NewInt(50), // TVL:   ~3M | 1M-15M+  | 50% RL
		"juno-1":       sdkmath.NewInt(50), // TVL:   ~3M | 1M-15M+  | 50% RL
		"osmosis-1":    sdkmath.NewInt(10), // TVL:  ~30M |    30M+  | 10% RL
		"phoenix-1":    sdkmath.ZeroInt(),  // TVL: ~190k |     <1M  | No rate limit
		"sommelier-3":  sdkmath.ZeroInt(),  // TVL: ~450k |     <1M  | No rate limit
		"stargaze-1":   sdkmath.NewInt(50), // TVL: 1.35M | 1M-15M+  | 50% RL
		"umee-1":       sdkmath.ZeroInt(),  // TVL: ~200k |     <1M  | No rate limit
	}

	// Osmo transfer channel is required for new rate limits
	OsmosisTransferChannelId = "channel-5"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	distributionkeeper distributionkeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v17...")

		ctx.Logger().Info("Migrating host zones...")
		if err := RegisterCommunityPoolAddresses(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate host zones")
		}

		ctx.Logger().Info("Deleting all pending queries...")
		DeleteAllStaleQueries(ctx, icqKeeper)

		ctx.Logger().Info("Reseting slash query in progress...")
		ResetSlashQueryInProgress(ctx, stakeibcKeeper)

		ctx.Logger().Info("Updating community pool tax...")
		if err := IncreaseCommunityPoolTax(ctx, distributionkeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to increase community pool tax")
		}

		ctx.Logger().Info("Updating redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Update rate limits thresholds...")
		UpdateRateLimitThresholds(ctx, stakeibcKeeper, ratelimitKeeper)

		ctx.Logger().Info("Adding rate limits to Osmosis...")
		if err := AddRateLimitToOsmosis(ctx, ratelimitKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to add rate limits to Osmosis")
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Migrates the host zones to the new structure which supports community pool liquid staking
// We don't have to perform a true migration here since only new fields were added
// (in other words, we can deserialize the old host zone structs into the new types)
// This will also register the relevant community pool ICA addresses
func RegisterCommunityPoolAddresses(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		chainId := hostZone.ChainId

		// Create and store a new community pool stake and redeem module address
		stakeHoldingAddress := stakeibctypes.NewHostZoneModuleAddress(
			chainId,
			stakeibckeeper.CommunityPoolStakeHoldingAddressKey,
		)
		redeemHoldingAddress := stakeibctypes.NewHostZoneModuleAddress(
			chainId,
			stakeibckeeper.CommunityPoolRedeemHoldingAddressKey,
		)

		if err := utils.CreateModuleAccount(ctx, k.AccountKeeper, stakeHoldingAddress); err != nil {
			return errorsmod.Wrapf(err, "unable to create community pool stake account for host zone %s", chainId)
		}
		if err := utils.CreateModuleAccount(ctx, k.AccountKeeper, redeemHoldingAddress); err != nil {
			return errorsmod.Wrapf(err, "unable to create community pool redeem account for host zone %s", chainId)
		}

		hostZone.CommunityPoolStakeHoldingAddress = stakeHoldingAddress.String()
		hostZone.CommunityPoolRedeemHoldingAddress = redeemHoldingAddress.String()

		k.SetHostZone(ctx, hostZone)

		// Register the deposit and return ICA addresses
		// (these will get set in the OnChanAck callback)
		// create community pool deposit account
		connectionId := hostZone.ConnectionId
		connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionId)
		if !found {
			return errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection %s not found", connectionId)
		}
		counterpartyConnectionId := connectionEnd.Counterparty.ConnectionId

		appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
			Version:                icatypes.Version,
			ControllerConnectionId: connectionId,
			HostConnectionId:       counterpartyConnectionId,
			Encoding:               icatypes.EncodingProtobuf,
			TxType:                 icatypes.TxTypeSDKMultiMsg,
		}))

		depositAccount := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, connectionId, depositAccount, appVersion); err != nil {
			return errorsmod.Wrapf(stakeibctypes.ErrFailedToRegisterHostZone, "failed to register community pool deposit ICA")
		}

		returnAccount := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_COMMUNITY_POOL_RETURN)
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, connectionId, returnAccount, appVersion); err != nil {
			return errorsmod.Wrapf(stakeibctypes.ErrFailedToRegisterHostZone, "failed to register community pool return ICA")
		}
	}

	return nil
}

// Deletes all stale queries
func DeleteAllStaleQueries(ctx sdk.Context, k icqkeeper.Keeper) {
	for _, query := range k.AllQueries(ctx) {
		if query.CallbackId == stakeibckeeper.ICQCallbackID_Delegation {
			k.DeleteQuery(ctx, query.Id)
		}
	}
}

// Resets the slash query in progress flag for each validator
func ResetSlashQueryInProgress(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		for i, validator := range hostZone.Validators {
			validator.SlashQueryInProgress = false
			hostZone.Validators[i] = validator
		}
		k.SetHostZone(ctx, hostZone)
	}
}

// Increases the community pool tax from 2 to 5%
// This was from prop 223 which passed, but was deleted due to an ICS blacklist
func IncreaseCommunityPoolTax(ctx sdk.Context, k distributionkeeper.Keeper) error {
	params := k.GetParams(ctx)
	params.CommunityTax = CommunityPoolTax
	return k.SetParams(ctx, params)
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		// Give osmosis a bit more slack since OSMO stakers collect real yield
		outerAdjustment := RedemptionRateOuterMaxAdjustment
		if hostZone.ChainId == OsmosisChainId {
			outerAdjustment = outerAdjustment.Add(OsmosisRedemptionRateBuffer)
		}

		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		innerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateInnerMinAdjustment)
		innerMaxDelta := hostZone.RedemptionRate.Mul(RedemptionRateInnerMaxAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(outerAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		innerMin := hostZone.RedemptionRate.Sub(innerMinDelta)
		innerMax := hostZone.RedemptionRate.Add(innerMaxDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MinInnerRedemptionRate = innerMin
		hostZone.MaxInnerRedemptionRate = innerMax
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}

// Update rate limits based on current TVL
func UpdateRateLimitThresholds(ctx sdk.Context, sk stakeibckeeper.Keeper, rk ratelimitkeeper.Keeper) {
	for _, rateLimit := range rk.GetAllRateLimits(ctx) {
		stDenom := rateLimit.Path.Denom
		hostDenom := stDenom[2:]

		// Lookup the associated host zone to get the chain ID
		hostZone, err := sk.GetHostZoneFromHostDenom(ctx, hostDenom)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("host zone not found for denom %s", hostDenom))
			continue
		}

		// Determine the expected rate limit threshold for the chain
		updatedThreshold, ok := UpdatedRateLimits[hostZone.ChainId]
		if !ok {
			ctx.Logger().Error(fmt.Sprintf("rate limit not specified for %s", hostZone.ChainId))
			continue
		}

		// If the expected threshold is 0, that means there should be no rate limit
		// Remove the rate limit in this case
		if updatedThreshold.IsZero() {
			rk.RemoveRateLimit(ctx, rateLimit.Path.Denom, rateLimit.Path.ChannelId)
			continue
		}

		rateLimit.Quota.MaxPercentRecv = updatedThreshold
		rateLimit.Quota.MaxPercentSend = updatedThreshold
		rk.SetRateLimit(ctx, rateLimit)
	}
}

// Rate limits transfers to osmosis across each stToken
func AddRateLimitToOsmosis(ctx sdk.Context, k ratelimitkeeper.Keeper) error {
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		denom := rateLimit.Path.Denom

		channelValue := k.GetChannelValue(ctx, denom)
		if channelValue.IsZero() {
			return ratelimittypes.ErrZeroChannelValue
		}

		// Ignore the rate limit if it already exists (e.g. stuosmo)
		_, found := k.GetRateLimit(ctx, rateLimit.Path.Denom, OsmosisTransferChannelId)
		if found {
			continue
		}

		// Create and store the rate limit object with the same bounds as
		// the original rate limit
		path := ratelimittypes.Path{
			Denom:     denom,
			ChannelId: OsmosisTransferChannelId,
		}
		quota := ratelimittypes.Quota{
			MaxPercentSend: rateLimit.Quota.MaxPercentSend,
			MaxPercentRecv: rateLimit.Quota.MaxPercentRecv,
			DurationHours:  rateLimit.Quota.DurationHours,
		}
		flow := ratelimittypes.Flow{
			Inflow:       sdkmath.ZeroInt(),
			Outflow:      sdkmath.ZeroInt(),
			ChannelValue: channelValue,
		}

		k.SetRateLimit(ctx, ratelimittypes.RateLimit{
			Path:  &path,
			Quota: &quota,
			Flow:  &flow,
		})
	}

	return nil
}
