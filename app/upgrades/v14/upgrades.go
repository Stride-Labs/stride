package v14

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v3/x/ccv/consumer/keeper"
	evmosvestingkeeper "github.com/evmos/vesting/x/vesting/keeper"
	"github.com/evmos/vesting/x/vesting/types"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"

	"github.com/Stride-Labs/stride/v14/utils"
	claimkeeper "github.com/Stride-Labs/stride/v14/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v14/x/claim/types"
	icqkeeper "github.com/Stride-Labs/stride/v14/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	stakeibcmigration "github.com/Stride-Labs/stride/v14/x/stakeibc/migrations/v3"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

var (
	UpgradeName = "v14"

	GaiaChainId = "cosmoshub-4"

	// Vesting
	VestingStartTimeAccount1 = int64(1662350400) // Sept 4, 2022
	VestingEndTimeAccount1   = int64(1788512452)
	VestingStartTimeAccount2 = int64(1662350400) // Sept 4, 2022
	VestingEndTimeAccount2   = int64(1820016452)
	Account1                 = "stride12z83xmrkr7stjk4q2vn95c02n7jryj55gd3aq3"
	Account1VestingUstrd     = int64(187_500_000_000)
	Account2                 = "stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg"
	Account2VestingUstrd     = int64(375_000_000_000)
	FunderAddress            = "stride1avdulp2p7jjv37valeyt4c6fn6qtfhevr2ej3r"

	// ICS
	DistributionTransmissionChannel = "channel-0"
	// Module account address for consumer_rewards_pool (see: https://github.com/cosmos/interchain-security/blob/main/x/ccv/provider/types/keys.go#L33C25-L33C46)
	ProviderFeePoolAddrStr         = "cosmos1ap0mh6xzfn8943urr84q6ae7zfnar48am2erhd"
	ConsumerRedistributionFraction = "0.85"
	Enabled                        = true
	RefundFraction                 = "0.4"
	// strided q auth module-account cons_to_send_to_provider
	ConsToSendToProvider = "stride1ywtansy6ss0jtq8ckrcv6jzkps8yh8mfmvxqvv"
	FeeCollector         = "stride17xpfvakm2amg962yls6f84z3kell8c5lnjrul3"

	// Airdrop params
	AirdropDuration  = time.Hour * 24 * 30 * 12 * 3                 // 3 years
	AirdropStartTime = time.Date(2023, 9, 4, 16, 0, 0, 0, time.UTC) // Sept 4, 2023 @ 16:00 UTC (12:00 EST)

	InjectiveAirdropDistributor = "stride1gxy4qnm7pg2wzfpc3j7rk7ggvyq2ls944f0wus"
	InjectiveAirdropIdentifier  = "injective"
	InjectiveChainId            = "injective-1"

	ComdexAirdropDistributor = "stride1quag8me3n7h7qw2z0fm7khdemwummep6lnn3ja"
	ComdexAirdropIdentifier  = "comdex"
	ComdexChainId            = "comdex-1"

	SommAirdropDistributor = "stride13xxegkegnezayceeqdy98v2k8xyat5ah4umdwk"
	SommAirdropIdentifier  = "sommelier"
	SommChainId            = "sommelier-3"

	UmeeAirdropDistributor = "stride1qkj9hh08zk44zrw2krv5vn34qn8cwt7h2ppfxu"
	UmeeAirdropIdentifier  = "umee"
	UmeeChainId            = "umee-1"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v14
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	claimKeeper claimkeeper.Keeper,
	consumerKeeper *ccvconsumerkeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	evmosvestingKeeper evmosvestingkeeper.Keeper,
	stakeibcStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v14...")

		evk := evmosvestingKeeper
		sk := stakingKeeper
		ak := accountKeeper
		bk := bankKeeper
		ck := consumerKeeper
		sibc := stakeibcKeeper
		currentVersions := mm.GetVersionMap()

		// Migrate the Validator and HostZone structs from stakeibc
		utils.LogModuleMigration(ctx, currentVersions, stakeibctypes.ModuleName)
		if err := stakeibcmigration.MigrateStore(ctx, stakeibcStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate stakeibc store")
		}

		// Update Stakeibc Params
		MigrateStakeibcParams(ctx, stakeibcKeeper)

		// Clear out any pending queries since the Query type updated
		// There shouldn't be any queries here unless the upgrade happened right at the epoch
		ClearPendingQueries(ctx, icqKeeper)

		// Enable LSM for the Gaia
		err := EnableLSMForGaia(ctx, stakeibcKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "unable to enable LSM for Gaia")
		}

		// Add airdrops for Injective, Comedex, Somm, and Umee
		if err := InitAirdrops(ctx, claimKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate airdrop")
		}

		// VESTING CHANGES
		// Migrate SL employee pool Account1 to evmos vesting account
		if err := MigrateAccount1(ctx, evk, sk, ak, bk); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate account 12z83x")
		}

		// Update vesting schedule - SL employee pool tokens were mistankenly assigned an investor vesting schedule
		// migrate Account2 from a ContinuousVestingAccount that starts on Sept 4, 2023 to a continuous vesting account that starts on Sept 4, 2022
		if err := MigrateAccount2(ctx, ak); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate account 1nwyvk")
		}

		// ICS CHANGES
		// In the v13 upgrade, params were reset to genesis. In v12, the version map wasn't updated. So when mm.RunMigrations(ctx, configurator, vm) ran
		// in v13, InitGenesis was run for ccvconsumer.
		if err := SetConsumerParams(ctx, ck, sibc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set consumer params")
		}
		// Since the last upgrade (which is also when rewards stopped accumulating), to much STRD has been sent to the consumer fee pool. This is because
		// ConsumerRedistributionFraction was updated from 0.85 to 0.75 in the last upgrade. 25% of inflation (instead of 15%) was being sent there. So,
		// we need to send 1-(15/25) = 40% of the STRD in the fee pool back to the fee distribution account.
		if err := SendConsumerFeePoolToFeeDistribution(ctx, ck, bk, ak, sk); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to send consumer fee pool to fee distribution")
		}

		// `RunMigrations` (below) checks the old consensus version of each module (found in
		// the store) and compares it against the updated consensus version in the binary
		// If the old and new consensus versions are not the same, it attempts to call that
		// module's migration function that must be registered ahead of time
		//
		// Since the migrations above were executed directly (instead of being registered
		// and invoked through a Migrator), we need to set the module versions in the versionMap
		// to the new version, to prevent RunMigrations from attempting to re-run each migrations
		vm[stakeibctypes.ModuleName] = currentVersions[stakeibctypes.ModuleName]

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func InitAirdrops(ctx sdk.Context, claimKeeper claimkeeper.Keeper) error {
	duration := uint64(AirdropDuration.Seconds())
	startTime := uint64(AirdropStartTime.Unix())

	// Add the Injective Airdrop
	ctx.Logger().Info("Adding Injective airdrop...")
	if err := claimKeeper.CreateAirdropAndEpoch(ctx, claimtypes.MsgCreateAirdrop{
		Distributor:      InjectiveAirdropDistributor,
		Identifier:       InjectiveAirdropIdentifier,
		ChainId:          InjectiveChainId,
		Denom:            claimtypes.DefaultClaimDenom,
		StartTime:        startTime,
		Duration:         duration,
		AutopilotEnabled: true,
	}); err != nil {
		return err
	}

	// Add the Comdex Airdrop
	ctx.Logger().Info("Adding Comdex airdrop...")
	if err := claimKeeper.CreateAirdropAndEpoch(ctx, claimtypes.MsgCreateAirdrop{
		Distributor:      ComdexAirdropDistributor,
		Identifier:       ComdexAirdropIdentifier,
		ChainId:          ComdexChainId,
		Denom:            claimtypes.DefaultClaimDenom,
		StartTime:        startTime,
		Duration:         duration,
		AutopilotEnabled: false,
	}); err != nil {
		return err
	}

	// Add the Somm Airdrop
	ctx.Logger().Info("Adding Somm airdrop...")
	if err := claimKeeper.CreateAirdropAndEpoch(ctx, claimtypes.MsgCreateAirdrop{
		Distributor:      SommAirdropDistributor,
		Identifier:       SommAirdropIdentifier,
		ChainId:          SommChainId,
		Denom:            claimtypes.DefaultClaimDenom,
		StartTime:        startTime,
		Duration:         duration,
		AutopilotEnabled: false,
	}); err != nil {
		return err
	}

	// Add the Umee Airdrop
	ctx.Logger().Info("Adding Umee airdrop...")
	if err := claimKeeper.CreateAirdropAndEpoch(ctx, claimtypes.MsgCreateAirdrop{
		Distributor:      UmeeAirdropDistributor,
		Identifier:       UmeeAirdropIdentifier,
		ChainId:          UmeeChainId,
		Denom:            claimtypes.DefaultClaimDenom,
		StartTime:        startTime,
		Duration:         duration,
		AutopilotEnabled: false,
	}); err != nil {
		return err
	}

	ctx.Logger().Info("Loading airdrop allocations...")
	claimKeeper.LoadAllocationData(ctx, allocations)
	return nil
}

func MigrateAccount1(ctx sdk.Context, evk evmosvestingkeeper.Keeper, sk stakingkeeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper) error {
	// fetch the account
	account := ak.GetAccount(ctx, sdk.MustAccAddressFromBech32(Account1))
	if account == nil {
		// account must be initialized
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s not found", Account1)
	}
	// First, reset the account as a base account. Only accounts that conform to the BaseAccount interface can be converted to ClawbackVestingAccounts
	baseVestingAcc := account.(*vesting.ContinuousVestingAccount).BaseVestingAccount
	baseAcc := baseVestingAcc.BaseAccount
	ak.SetAccount(ctx, baseAcc)

	// Then, create the clawback vesting account. This will reset the account type
	createClawbackMsg := &types.MsgCreateClawbackVestingAccount{
		FunderAddress:     FunderAddress,
		VestingAddress:    Account1,
		EnableGovClawback: false,
	}
	_, err := evk.CreateClawbackVestingAccount(sdk.WrapSDKContext(ctx), createClawbackMsg)
	if err != nil {
		return err
	}

	// TODO: verify sk.BondDenom(ctx) is ustrd by querying the account after running localstride
	// NOTE: LockupPeriods adds a transfer restriction. Unvested tokens are also transfer restricted. The behavior we want is
	// Vested:
	// - transferable
	// - delegatable
	// - votable
	// Unvested:
	// - not transferable
	// - delegatable
	// - votable
	// This is the default behavior (without a lockup). So, we don't add LockupPeriods.
	fundAccMsg := &types.MsgFundVestingAccount{
		FunderAddress:  FunderAddress,
		VestingAddress: Account1,
		StartTime:      time.Unix(VestingStartTimeAccount1, 0),
		VestingPeriods: sdkvesting.Periods{
			// Period is 3 years
			// 60*60*24*365*3 seconds
			{Length: 94608000, Amount: sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), sdk.NewInt(Account1VestingUstrd)))},
		},
	}

	// Then, fund the account
	err = FundVestingAccount(ctx, evk, sk, ak, bk, fundAccMsg)
	if err != nil {
		return err
	}

	return nil
}

// Migrate the stakeibc params, specifically:
//   - Remove SafetyNumValidators
//   - Remove SafetyMaxSlashPercentage
//   - Add ValidatorSlashQueryThreshold
//
// NOTE: If a parameter is added, the old params cannot be unmarshalled
// to the new schema. To get around this, we have to set each parameter explicitly
// Considering all mainnet stakeibc params are set to the default, we can just use that
func MigrateStakeibcParams(ctx sdk.Context, k stakeibckeeper.Keeper) {
	params := stakeibctypes.DefaultParams()
	k.SetParams(ctx, params)
}

// Since the Query struct was updated, it's easier to just clear out any pending
// queries rather than attempt to migrate them
func ClearPendingQueries(ctx sdk.Context, k icqkeeper.Keeper) {
	for _, query := range k.AllQueries(ctx) {
		k.DeleteQuery(ctx, query.Id)
	}
}

// Enable LSM liquid stakes for Gaia
func EnableLSMForGaia(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	hostZone, found := k.GetHostZone(ctx, GaiaChainId)
	if !found {
		return stakeibctypes.ErrHostZoneNotFound.Wrapf(GaiaChainId)
	}

	hostZone.LsmLiquidStakeEnabled = true
	k.SetHostZone(ctx, hostZone)

	return nil
}
func MigrateAccount2(ctx sdk.Context, ak authkeeper.AccountKeeper) error {
	// Get account
	account := ak.GetAccount(ctx, sdk.MustAccAddressFromBech32(Account2))
	if account == nil {
		return nil
	}
	// change the start_time to Sept 4, 2022. The ugprade goes live on or after Sept 4, 2023, so the first year vest is still enforced
	// (the account was previously set to start on Sept 4, 2023)
	account.(*vesting.ContinuousVestingAccount).StartTime = VestingStartTimeAccount2
	// NOTE: we shouldn't have to update delegated_vesting on the BaseAccount. That's because,
	// DF (delegated free) and DV (delegated vesting) coins are set on (un)delegation and are point-in-time.
	// So, delegated_vesting overcounts how many tokens are vesting. Whenever an undelegation occurs, DF and DV should be set correctly.
	// See: https://github.com/cosmos/cosmos-sdk/commit/c5238b0d1ecfef8be3ccdaee02d23ee93ef9c69b
	// set the account
	ak.SetAccount(ctx, account)
	return nil
}

func SetConsumerParams(ctx sdk.Context, ck *ccvconsumerkeeper.Keeper, sibc stakeibckeeper.Keeper) error {

	// Pre-upgrade params
	// "params": {
	// 		"enabled": false, Set to true
	// 		"blocks_per_distribution_transmission": "1000", OK
	// 		"distribution_transmission_channel": "", Set to channel-0
	//	 	"provider_fee_pool_addr_str": "", Set to address
	//	 	"ccv_timeout_period": "2419200s", OK
	//	 	"transfer_timeout_period": "3600s", OK
	//	 	"consumer_redistribution_fraction": "0.75", Set to 0.85
	//	 	"historical_entries": "10000",  OK
	//	 	"unbonding_period": "1728000s", reset in proposal:
	//	 	"soft_opt_out_threshold": "0.05", OK
	//	 	"reward_denoms": [], Set to stTokens and revert change
	//	 	"provider_reward_denoms": [] Leave unset
	// 	}
	// Params should match https://dev.mintscan.io/cosmos/proposals/799

	ccvconsumerparams := ck.GetConsumerParams(ctx)
	ccvconsumerparams.Enabled = true
	ccvconsumerparams.DistributionTransmissionChannel = DistributionTransmissionChannel
	ccvconsumerparams.ProviderFeePoolAddrStr = ProviderFeePoolAddrStr
	ccvconsumerparams.ConsumerRedistributionFraction = ConsumerRedistributionFraction
	ck.SetParams(ctx, ccvconsumerparams)

	// Then, add the stTokens to the reward list
	ctx.Logger().Info("Registering stTokens to consumer reward denom whitelist...")
	hostZones := sibc.GetAllHostZone(ctx)
	allDenoms := []string{}

	// get all stToken denoms
	for _, zone := range hostZones {
		allDenoms = append(allDenoms, stakeibctypes.StAssetDenomFromHostZoneDenom(zone.HostDenom))
	}

	err := sibc.RegisterStTokenDenomsToWhitelist(ctx, allDenoms)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to register stTokens to whitelist")
	}

	return nil
}

func SendConsumerFeePoolToFeeDistribution(ctx sdk.Context, ck *ccvconsumerkeeper.Keeper, bk bankkeeper.Keeper, ak authkeeper.AccountKeeper, sk stakingkeeper.Keeper) error {
	// Read account balance of consumer fee account
	address := sdk.MustAccAddressFromBech32(ConsToSendToProvider)
	frac, err := sdk.NewDecFromStr(RefundFraction)
	if err != nil {
		// ConsumerRedistributionFrac was already validated when set as a param
		panic(fmt.Errorf("ConsumerRedistributionFrac is invalid: %w", err))
	}

	total := bk.GetBalance(ctx, address, sk.BondDenom(ctx))
	totalTokens := sdk.NewDecCoinsFromCoins(total)
	// truncated decimals are implicitly added to provider
	refundTokens, _ := totalTokens.MulDec(frac).TruncateDecimal()
	for _, token := range refundTokens {
		// Send tokens back to the fee distributinon address
		// NOTE: This is technically not a module account because it's removed from maccPerms (in order to allow ibc transfers, see: https://github.com/cosmos/ibc-go/issues/1889)
		// But conceptually it's a module account
		err := bk.SendCoinsFromAccountToModule(ctx, address, authtypes.FeeCollectorName, sdk.NewCoins(token))
		if err != nil {
			return errorsmod.Wrapf(err, "unable to send consumer fee pool to fee distribution")
		}
	}
	return nil
}

// ---------------------------- Evmos vesting logic ------------------------------------
// NOTE: This is mostly copy+pasted from Evmos vesting module
// However, some of the functions were private so couldn't be used directly and the token transfer (funding) logic was removed

// FundVestingAccount funds a ClawbackVestingAccount with the provided amount.
// This can only be executed by the funder of the vesting account.
//
// Checks performed on the ValidateBasic include:
//   - funder and vesting addresses are correct bech32 format
//   - vesting address is not the zero address
//   - both vesting and lockup periods are non-empty
//   - both lockup and vesting periods contain valid amounts and lengths
//   - both vesting and lockup periods describe the same total amount
func FundVestingAccount(ctx sdk.Context, k evmosvestingkeeper.Keeper, stakingKeeper stakingkeeper.Keeper, accountKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.Keeper, msg *types.MsgFundVestingAccount) error {
	ak := accountKeeper
	bk := bankKeeper

	// Error checked during msg validation
	// CHANGE: funderAddr isn't used because we're doing a migration
	// funderAddr := sdk.MustAccAddressFromBech32(msg.FunderAddress)
	vestingAddr := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	if bk.BlockedAddr(vestingAddr) {
		return errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"%s is not allowed to receive funds", msg.VestingAddress,
		)
	}

	// Check if vesting account exists
	vestingAcc, err := k.GetClawbackVestingAccount(ctx, vestingAddr)
	if err != nil {
		return err
	}

	vestingCoins := msg.VestingPeriods.TotalAmount()
	lockupCoins := msg.LockupPeriods.TotalAmount()

	fmt.Println("vestingCoins: ", vestingCoins)
	fmt.Println("lockupCoins: ", lockupCoins)

	// If lockup absent, default to an instant unlock schedule
	if !vestingCoins.IsZero() && len(msg.LockupPeriods) == 0 {
		msg.LockupPeriods = sdkvesting.Periods{
			{Length: 0, Amount: vestingCoins},
		}
		lockupCoins = vestingCoins
	}

	// If vesting absent, default to an instant vesting schedule
	if !lockupCoins.IsZero() && len(msg.VestingPeriods) == 0 {
		msg.VestingPeriods = sdkvesting.Periods{
			{Length: 0, Amount: lockupCoins},
		}
		vestingCoins = lockupCoins
	}

	if msg.FunderAddress != vestingAcc.FunderAddress {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s can only accept grants from account %s", msg.VestingAddress, vestingAcc.FunderAddress)
	}

	// CHANGE: redefine addGrant below and pass in the vesting/staking keepers
	err = addGrant(ctx, k, stakingKeeper, vestingAcc, msg.GetStartTime().Unix(), msg.GetLockupPeriods(), msg.GetVestingPeriods(), vestingCoins)
	if err != nil {
		return err
	}
	ak.SetAccount(ctx, vestingAcc)

	// CHANGE: because we're doing a migration, we don't need to send coins from the funder to the vesting account
	// Send coins from the funder to vesting account
	// if err = bk.SendCoins(ctx, funderAddr, vestingAddr, vestingCoins); err != nil {
	// 	return nil, err
	// }

	telemetry.IncrCounter(
		float32(ctx.GasMeter().GasConsumed()),
		"tx", "fund_vesting_account", "gas_used",
	)
	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				evmosvestingtypes.EventTypeFundVestingAccount,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.FunderAddress),
				sdk.NewAttribute(evmosvestingtypes.AttributeKeyCoins, vestingCoins.String()),
				sdk.NewAttribute(evmosvestingtypes.AttributeKeyStartTime, msg.StartTime.String()),
				sdk.NewAttribute(evmosvestingtypes.AttributeKeyAccount, msg.VestingAddress),
			),
		},
	)

	return nil
}

// addGrant merges a new clawback vesting grant into an existing
// ClawbackVestingAccount.
func addGrant(
	ctx sdk.Context,
	k evmosvestingkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	va *types.ClawbackVestingAccount,
	grantStartTime int64,
	grantLockupPeriods, grantVestingPeriods sdkvesting.Periods,
	grantCoins sdk.Coins,
) error {
	// check if the clawback vesting account has only been initialized and not yet funded --
	// in that case it's necessary to update the vesting account with the given start time because this is set to zero in the initialization
	if len(va.LockupPeriods) == 0 && len(va.VestingPeriods) == 0 {
		va.StartTime = time.Unix(grantStartTime, 0).UTC()
	}

	// how much is really delegated?
	bondedAmt := stakingKeeper.GetDelegatorBonded(ctx, va.GetAddress())
	unbondingAmt := stakingKeeper.GetDelegatorUnbonding(ctx, va.GetAddress())
	delegatedAmt := bondedAmt.Add(unbondingAmt)
	delegated := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), delegatedAmt))

	// modify schedules for the new grant
	newLockupStart, newLockupEnd, newLockupPeriods := types.DisjunctPeriods(va.GetStartTime(), grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := types.DisjunctPeriods(
		va.GetStartTime(),
		grantStartTime,
		va.GetVestingPeriods(),
		grantVestingPeriods,
	)

	if newLockupStart != newVestingStart {
		return errorsmod.Wrapf(
			types.ErrVestingLockup,
			"vesting start time calculation should match lockup start (%d ≠ %d)",
			newVestingStart, newLockupStart,
		)
	}

	va.StartTime = time.Unix(newLockupStart, 0).UTC()
	va.EndTime = types.Max64(newLockupEnd, newVestingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	va.OriginalVesting = va.OriginalVesting.Add(grantCoins...)

	// cap DV at the current unvested amount, DF rounds out to current delegated
	unvested := va.GetVestingCoins(ctx.BlockTime())
	va.DelegatedVesting = delegated.Min(unvested)
	va.DelegatedFree = delegated.Sub(va.DelegatedVesting...)
	return nil
}
