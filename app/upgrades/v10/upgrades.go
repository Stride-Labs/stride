package v10

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icacontrollermigrations "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	clientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"

	claimkeeper "github.com/Stride-Labs/stride/v14/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v14/x/claim/types"
	icacallbackskeeper "github.com/Stride-Labs/stride/v14/x/icacallbacks/keeper"
	mintkeeper "github.com/Stride-Labs/stride/v14/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v14/x/mint/types"
	ratelimitkeeper "github.com/Stride-Labs/stride/v14/x/ratelimit/keeper"
	ratelimitgov "github.com/Stride-Labs/stride/v14/x/ratelimit/keeper/gov"
	ratelimittypes "github.com/Stride-Labs/stride/v14/x/ratelimit/types"
	recordskeeper "github.com/Stride-Labs/stride/v14/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	deprecatedproto "github.com/golang/protobuf/proto" //nolint:staticcheck
)

var (
	UpgradeName     = "v10"
	EpochProvisions = sdk.NewDec(929_681_506)

	StakingProportion                     = "0.1603620"
	CommunityPoolGrowthProportion         = "0.2158275"
	StrategicReserveProportion            = "0.4879320"
	CommunityPoolSecurityBudgetProportion = "0.1358785"

	CommunityPoolGrowthAddress = "stride1lj0m72d70qerts9ksrsphy9nmsd4h0s88ll9gfphmhemh8ewet5qj44jc9"
	BadKidsCustodian           = "stride17z6yy8vfgklgej9m848jm7rkp270gd9pgaw8zu"
	BadKidsTransferAmount      = sdk.NewInt(15_000_000_000)
	Ustrd                      = "ustrd"

	MinInitialDepositRatio = "0.50"

	// airdrop distributor addresses
	DistributorAddresses = map[string]string{
		"stride":  "stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev",
		"gaia":    "stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz",
		"osmosis": "stride1zlu2l3lx5tqvzspvjwsw9u0e907kelhqae3yhk",
		"juno":    "stride14k9g9zpgaycpey9840nnpa66l4nd6lu7g7t74c",
		"stars":   "stride12pum4adk5dhp32d90f8g8gfwujm0gwxqnrdlum",
		"evmos":   "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j",
	}
	NewDistributorAddresses = map[string]string{
		"stride":  "stride1w02dg74j8s38gqn6mvlr87hkvyv5rgp3cqe9se",
		"gaia":    "stride1w0w0gr6u796y2mjl9fuqt66jqvk3j59jq3jtpg",
		"osmosis": "stride1mfg5ck02tlyzdtdpaj70ngtgjs2vuawtkfz7xd",
		"juno":    "stride1ral4dsqk0nzyqlwtuyxgavfvx8hegml7u0rzx3",
		"stars":   "stride1rm9nxc5pw3k5r5s6lm85k73mfp734nhnxq570g",
		"evmos":   "stride1ej4e7x2hanmy6vrzrjh06g6dnfq5kxm73dmgsw",
	}

	RateLimitDurationHours = uint64(24)
	NewRateLimits          = map[string]sdkmath.Int{
		"cosmoshub-4":  sdkmath.NewInt(25),
		"osmosis-1":    sdkmath.NewInt(25),
		"injective-1":  sdkmath.NewInt(25),
		"evmos_9001-2": sdkmath.NewInt(25),
		"juno-1":       sdkmath.NewInt(75),
		"stargaze-1":   sdkmath.NewInt(75),
		"phoenix-1":    sdkmath.NewInt(75),
		"umee-1":       sdkmath.NewInt(75),
	}
)

// CreateUpgradeHandler creates an SDK upgrade handler for v10
// This upgrade:
// - Upgrades the SDK from v0.46 to v0.47
// - Upgrades ibc-go from v5 to v7.1
// - Reduces STRD staking rewards
// - Executes Prop #205
// - Enables rate limits
// - Migrates the airdrop distributor address
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	capabilityStoreKey *storetypes.KVStoreKey,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	capabilityKeeper *capabilitykeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	claimKeeper claimkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	consensusParamsKeeper consensusparamkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	icacallbacksKeeper icacallbackskeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
	paramsKeeper paramskeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		ctx.Logger().Info("Starting upgrade v10...")

		ctx.Logger().Info("Migrating tendermint consensus params from x/params to x/consensus...")
		legacyParamSubspace := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		baseapp.MigrateParams(ctx, legacyParamSubspace, &consensusParamsKeeper)

		ctx.Logger().Info("Migrating ICA channel capabilities for ibc-go v5 to v6 migration...")
		if err := icacontrollermigrations.MigrateICS27ChannelCapability(
			ctx,
			cdc,
			capabilityStoreKey,
			capabilityKeeper,
			stakeibctypes.ModuleName,
		); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to migrate ICA channel capabilities")
		}

		ctx.Logger().Info("Adding localhost IBC client for ibc-go v7.0 to v7.1 migration...")
		AddLocalhostIBCClient(ctx, clientKeeper)

		ctx.Logger().Info("Pruning expired tendermint consensus states...")
		if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, cdc, clientKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to prune expired consensus states")
		}

		ctx.Logger().Info("Reducing STRD staking rewards...")
		if err := ReduceSTRDStakingRewards(ctx, mintKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to reduce STRD staking rewards")
		}

		ctx.Logger().Info("Migrating callback data...")
		if err := MigrateCallbackData(ctx, icacallbacksKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to migrate callback data")
		}

		ctx.Logger().Info("Running module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)

		ctx.Logger().Info("Setting MinInitialDepositRatio...")
		if err := SetMinInitialDepositRatio(ctx, govKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to set MinInitialDepositRatio")
		}

		ctx.Logger().Info("Migrating claim distributor addresses...")
		if err := MigrateClaimDistributorAddress(ctx, claimKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to MigrateClaimDistributorAddress")
		}

		ctx.Logger().Info("Executing Prop #205...")
		if err := ExecuteProp205(ctx, bankKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to submit transfer for Prop #205")
		}

		ctx.Logger().Info("Enabling rate limits...")
		if err := EnableRateLimits(ctx, accountKeeper, channelKeeper, ratelimitKeeper, stakeibcKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to enable rate limits")
		}

		ctx.Logger().Info("v10 Upgrade Complete")
		return vm, err
	}
}

// Cut STRD staking rewards in half (staking rewards make up 27.64% of total provisions)
// Reduce epoch provisions by 13.82% from 1,078,767,123 to 929,681,506
func ReduceSTRDStakingRewards(ctx sdk.Context, k mintkeeper.Keeper) error {
	minter := minttypes.NewMinter(EpochProvisions)
	k.SetMinter(ctx, minter)

	stakingProportion := sdk.MustNewDecFromStr(StakingProportion)
	communityPoolGrowthProportion := sdk.MustNewDecFromStr(CommunityPoolGrowthProportion)
	strategicReserveProportion := sdk.MustNewDecFromStr(StrategicReserveProportion)
	communityPoolSecurityBudgetProportion := sdk.MustNewDecFromStr(CommunityPoolSecurityBudgetProportion)

	// Confirm proportions sum to 100
	totalProportions := stakingProportion.
		Add(communityPoolGrowthProportion).
		Add(strategicReserveProportion).
		Add(communityPoolSecurityBudgetProportion)

	if !totalProportions.Equal(sdk.OneDec()) {
		return fmt.Errorf("distribution proportions do not sum to 1 (%v)", totalProportions)
	}

	distributionProperties := minttypes.DistributionProportions{
		Staking:                     stakingProportion,
		CommunityPoolGrowth:         communityPoolGrowthProportion,
		StrategicReserve:            strategicReserveProportion,
		CommunityPoolSecurityBudget: communityPoolSecurityBudgetProportion,
	}

	params := k.GetParams(ctx)
	params.DistributionProportions = distributionProperties
	k.SetParams(ctx, params)

	return nil
}

// Set the initial deposit ratio to 25%
func SetMinInitialDepositRatio(ctx sdk.Context, k govkeeper.Keeper) error {
	params := k.GetParams(ctx)
	params.MinInitialDepositRatio = MinInitialDepositRatio
	return k.SetParams(ctx, params)
}

// This likely isn't necessary, but since migrating from google proto to
// cosmos gogoproto has the potential for serialization differences,
// this reserializes all the callbacks with cosmos proto
func MigrateCallbackData(ctx sdk.Context, k icacallbackskeeper.Keeper) error {
	for _, oldCallbackData := range k.GetAllCallbackData(ctx) {
		oldCallbackArgsBz := oldCallbackData.CallbackArgs

		var newCallbackArgsBz []byte
		var err error

		switch oldCallbackData.CallbackId {
		case stakeibckeeper.ICACallbackID_Claim:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &stakeibctypes.ClaimCallback{})
		case stakeibckeeper.ICACallbackID_Delegate:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &stakeibctypes.DelegateCallback{})
		case stakeibckeeper.ICACallbackID_Rebalance:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &stakeibctypes.RebalanceCallback{})
		case stakeibckeeper.ICACallbackID_Redemption:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &stakeibctypes.RedemptionCallback{})
		case stakeibckeeper.ICACallbackID_Reinvest:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &stakeibctypes.ReinvestCallback{})
		case stakeibckeeper.ICACallbackID_Undelegate:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &stakeibctypes.UndelegateCallback{})
		case recordskeeper.IBCCallbacksID_NativeTransfer:
			newCallbackArgsBz, err = reserializeCallback(oldCallbackArgsBz, &recordstypes.TransferCallback{})
		}
		if err != nil {
			return err
		}

		newCallbackData := oldCallbackData
		newCallbackData.CallbackArgs = newCallbackArgsBz
		k.SetCallbackData(ctx, newCallbackData)
	}
	return nil
}

// Migrate the claim distributor address, change nothing else about the airdrop params
func MigrateClaimDistributorAddress(ctx sdk.Context, ck claimkeeper.Keeper) error {
	// Migrate claim params
	claimParams, err := ck.GetParams(ctx)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get claim parameters")
	}

	updatedAirdrops := []*claimtypes.Airdrop{}
	for _, airdrop := range claimParams.Airdrops {
		airdrop.DistributorAddress = NewDistributorAddresses[airdrop.AirdropIdentifier]
		updatedAirdrops = append(updatedAirdrops, airdrop)
	}
	claimParams.Airdrops = updatedAirdrops
	return ck.SetParams(ctx, claimParams)
}

// Helper function to deserialize using the deprecated proto types and reserialize using the new proto types
func reserializeCallback(oldCallbackArgsBz []byte, callback deprecatedproto.Message) ([]byte, error) {
	if err := deprecatedproto.Unmarshal(oldCallbackArgsBz, callback); err != nil {
		return nil, err
	}
	newCallbackArgs, err := cosmosproto.Marshal(callback)
	if err != nil {
		return nil, err
	}
	return newCallbackArgs, nil
}

// Explicitly update the IBC 02-client params, adding the localhost client type
func AddLocalhostIBCClient(ctx sdk.Context, k clientkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.AllowedClients = append(params.AllowedClients, exported.Localhost)
	k.SetParams(ctx, params)
}

// Execute Prop #205 - Signaling proposal to acquire Bad Kids
// Sends 15,000 STRD from "Community Pool - Growth" to the bad kids custodian account
func ExecuteProp205(ctx sdk.Context, k bankkeeper.Keeper) error {
	communityPoolGrowthAddress := sdk.MustAccAddressFromBech32(CommunityPoolGrowthAddress)
	badKidsCustodianAddress := sdk.MustAccAddressFromBech32(BadKidsCustodian)
	transferCoin := sdk.NewCoin(Ustrd, BadKidsTransferAmount)
	return k.SendCoins(ctx, communityPoolGrowthAddress, badKidsCustodianAddress, sdk.NewCoins(transferCoin))
}

// Enable the following rate limits:
//
//	ATOM:  25%
//	OSMO:  25%
//	INJ:   25%
//	EVMOS: 25%
//	JUNO:  75%
//	STARS: 75%
//	LUNA:  75%
//	UMEE:  75%
func EnableRateLimits(
	ctx sdk.Context,
	accountKeeper authkeeper.AccountKeeper,
	channelKeeper channelkeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) error {
	for _, hostZone := range stakeibcKeeper.GetAllHostZone(ctx) {
		threshold, shouldAddRateLimit := NewRateLimits[hostZone.ChainId]
		if !shouldAddRateLimit {
			continue
		}

		denom := stakeibctypes.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
		channelId := hostZone.TransferChannelId

		addRateLimit := &ratelimittypes.AddRateLimitProposal{
			Denom:          denom,
			ChannelId:      channelId,
			MaxPercentSend: threshold,
			MaxPercentRecv: threshold,
			DurationHours:  RateLimitDurationHours,
		}

		if err := ratelimitgov.AddRateLimit(ctx, ratelimitKeeper, channelKeeper, addRateLimit); err != nil {
			return errorsmod.Wrapf(err, "unable to add rate limit for %s", denom)
		}

		if hostZone.DelegationIcaAddress == "" {
			return stakeibctypes.ErrICAAccountNotFound
		}
		if hostZone.FeeIcaAddress == "" {
			return stakeibctypes.ErrICAAccountNotFound
		}

		ratelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.DepositAddress,
			Receiver: hostZone.DelegationIcaAddress,
		})

		rewardCollectorAddress := accountKeeper.GetModuleAccount(ctx, stakeibctypes.RewardCollectorName).GetAddress()
		ratelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.FeeIcaAddress,
			Receiver: rewardCollectorAddress.String(),
		})
	}
	return nil
}
