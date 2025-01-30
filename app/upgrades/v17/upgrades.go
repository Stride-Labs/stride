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

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"

	ratelimitkeeper "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/keeper"

	"github.com/Stride-Labs/stride/v25/utils"
	icqkeeper "github.com/Stride-Labs/stride/v25/x/interchainquery/keeper"
	recordtypes "github.com/Stride-Labs/stride/v25/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v25/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v25/x/stakeibc/types"
)

var (
	UpgradeName = "v17"

	// Community pool tax updated from 2 -> 5%
	CommunityPoolTax = sdk.MustNewDecFromStr("0.05")

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")

	// Define the hub chainId for disabling tokenization
	GaiaChainId = "cosmoshub-4"

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdk.MustNewDecFromStr("0.02")

	// Rate limits updated according to TVL
	// Framework:
	//       < 2.5M: No rate limit
	//   2.5M - 10M: 50%
	//    10M - 20M: 25%
	//    20M - 40M: 20%
	//    40M - 50M: 15%
	//        > 50M: 10%
	UpdatedRateLimits = map[string]sdkmath.Int{
		"comdex-1":     sdkmath.ZeroInt(),  // TVL: ~150k |   <2.5M  | No rate limit
		"cosmoshub-4":  sdkmath.NewInt(15), // TVL:  ~45M |  40M-50M | 15% RL
		"evmos_9001-2": sdkmath.NewInt(50), // TVL:   ~3M | 2.5M-10M | 50% RL
		"injective-1":  sdkmath.ZeroInt(),  // TVL: ~1.5M |   <2.5M  | No rate limit
		"juno-1":       sdkmath.NewInt(50), // TVL:   ~3M | 2.5M-10M | 50% RL
		"osmosis-1":    sdkmath.NewInt(15), // TVL:  ~45M |  40M-50M | 15% RL
		"phoenix-1":    sdkmath.ZeroInt(),  // TVL: ~200k |   <2.5M  | No rate limit
		"sommelier-3":  sdkmath.ZeroInt(),  // TVL: ~500k |   <2.5M  | No rate limit
		"stargaze-1":   sdkmath.ZeroInt(),  // TVL:  1.5M |   <2.5M  | No rate limit
		"umee-1":       sdkmath.ZeroInt(),  // TVL: ~150k |   <2.5M  | No rate limit
	}

	// Osmo transfer channel is required for new rate limits
	OsmosisTransferChannelId = "channel-5"

	// Constants for Prop 225
	CommunityPoolGrowthAddress = "stride1lj0m72d70qerts9ksrsphy9nmsd4h0s88ll9gfphmhemh8ewet5qj44jc9"
	LiquidityReceiver          = "stride1auhjs4zgp3ahvrpkspf088r2psz7wpyrypcnal"
	Prop225TransferAmount      = sdk.NewInt(31_572_300_000)
	Ustrd                      = "ustrd"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v17
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
	distributionkeeper distributionkeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v17...")

		ctx.Logger().Info("Migrating stakeibc params...")
		MigrateStakeibcParams(ctx, stakeibcKeeper)

		ctx.Logger().Info("Migrating Unbonding Records...")
		if err := MigrateUnbondingRecords(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate unbonding records")
		}

		ctx.Logger().Info("Migrating host zones...")
		if err := RegisterCommunityPoolAddresses(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to register community pool addresses on host zones")
		}

		ctx.Logger().Info("Deleting all pending slash queries...")
		DeleteAllStaleQueries(ctx, icqKeeper)

		ctx.Logger().Info("Resetting slash query in progress...")
		ResetSlashQueryInProgress(ctx, stakeibcKeeper)

		ctx.Logger().Info("Updating community pool tax...")
		if err := ExecuteProp223(ctx, distributionkeeper); err != nil {
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

		ctx.Logger().Info("Executing Prop 225, SHD Liquidity")
		if err := ExecuteProp225(ctx, bankKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to execute prop 225")
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Migrate the stakeibc params to add the ValidatorWeightCap parameter
//
// NOTE: If a parameter is added, the old params cannot be unmarshalled
// to the new schema. To get around this, we have to set each parameter explicitly
// Considering all mainnet stakeibc params are set to the default, we can just use that
func MigrateStakeibcParams(ctx sdk.Context, k stakeibckeeper.Keeper) {
	params := stakeibctypes.DefaultParams()
	k.SetParams(ctx, params)
}

// Migrate the user redemption records to add the stToken amount, calculated by estimating
// the redemption rate from the corresponding host zone unbonding records
// UserUnbondingRecords previously only used Native Token Amounts, we now want to use StTokenAmounts
// We only really need to migrate records in status UNBONDING_QUEUE or UNBONDING_IN_PROGRESS
// because the stToken amount is never used after unbonding is initiated
func MigrateUnbondingRecords(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			// If a record is in state claimable, the native token amount can't be trusted
			// since it gets decremented with each claim
			// As a result, we can't accurately estimate the redemption rate for these
			// user redemption records (but it also doesn't matter since the stToken
			// amount on the records is not used)
			if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_CLAIMABLE {
				continue
			}
			// similarly, if there aren't any tokens to unbond, we don't want to modify the record
			// as we won't be able to estimate a redemption rate
			if hostZoneUnbonding.NativeTokenAmount.IsZero() {
				continue
			}

			// Calculate the estimated redemption rate
			nativeTokenAmountDec := sdk.NewDecFromInt(hostZoneUnbonding.NativeTokenAmount)
			stTokenAmountDec := sdk.NewDecFromInt(hostZoneUnbonding.StTokenAmount)
			// this estimated rate is the amount of stTokens that would be received for 1 native token
			// e.g. if the rate is 0.5, then 1 native token would be worth 0.5 stTokens
			// estimatedStTokenConversionRate is 1 / redemption rate
			estimatedStTokenConversionRate := stTokenAmountDec.Quo(nativeTokenAmountDec)

			// Loop through User Redemption Records and insert an estimated stTokenAmount
			for _, userRedemptionRecordId := range hostZoneUnbonding.UserRedemptionRecords {
				userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
				if !found {
					// this would happen if the user has already claimed the unbonding, but given the status check above, this should never happen
					k.Logger(ctx).Error(fmt.Sprintf("user redemption record %s not found", userRedemptionRecordId))
					continue
				}

				userRedemptionRecord.StTokenAmount = estimatedStTokenConversionRate.Mul(sdk.NewDecFromInt(userRedemptionRecord.NativeTokenAmount)).RoundInt()
				k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
			}
		}
	}

	return nil
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
func ExecuteProp223(ctx sdk.Context, k distributionkeeper.Keeper) error {
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
		outerMaxDelta := hostZone.RedemptionRate.Mul(outerAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
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

// Execute Prop 225, release STRD to stride1auhjs4zgp3ahvrpkspf088r2psz7wpyrypcnal
func ExecuteProp225(ctx sdk.Context, k bankkeeper.Keeper) error {
	communityPoolGrowthAddress := sdk.MustAccAddressFromBech32(CommunityPoolGrowthAddress)
	liquidityReceiverAddress := sdk.MustAccAddressFromBech32(LiquidityReceiver)
	transferCoin := sdk.NewCoin(Ustrd, Prop225TransferAmount)
	return k.SendCoins(ctx, communityPoolGrowthAddress, liquidityReceiverAddress, sdk.NewCoins(transferCoin))
}
