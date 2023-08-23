package v14

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v13/x/stakeibc/keeper"

	// evmosvestingclient "github.com/evmos/vesting/x/vesting/client"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmosvestingkeeper "github.com/evmos/vesting/x/vesting/keeper"
	"github.com/evmos/vesting/x/vesting/types"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"

	// "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	UpgradeName = "v14"

	VestingStartTime_12z83xmr = int64(1662350400) // Sept 4, 2022
	VestingStartTime_1nwyvkxm = int64(1662350400) // Sept 4, 2022
	Account1                  = "stride12z83xmrkr7stjk4q2vn95c02n7jryj55gd3aq3"
	Account2                  = "stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v14
func CreateUpgradeHandler(
	mm *module.Manager,
	evmosvestingKeeper evmosvestingkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v14...")
		evk := evmosvestingKeeper
		sk := stakingKeeper
		ak := accountKeeper
		bk := bankKeeper

		// Migrate SL employee pool Account1 to evmos vesting account
		if err := MigrateAccount1(ctx, evk, sk, ak, bk); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate account 12z83x")
		}

		// Update vesting schedule - SL employee pool tokens were mistankenly assigned an investor vesting schedule
		// migrate Account2 from a ContinuousVestingAccount that starts on Sept 4, 2023 to a continuous vesting account that starts on Sept 4, 2022
		if err := MigrateAccount2(ctx, ak); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate account 1nwyvk")
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func MigrateAccount1(ctx sdk.Context, evk evmosvestingkeeper.Keeper, sk stakingkeeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper) error {
	// fetch the account
	account := ak.GetAccount(ctx, sdk.AccAddress(Account1))
	if account == nil {
		return nil
	}
	// First, create the clawback vesting account. This will reset the account type
	createClawbackMsg := &types.MsgCreateClawbackVestingAccount{
		FunderAddress:     Account1,
		VestingAddress:    Account1,
		EnableGovClawback: false,
	}
	_, err := evk.CreateClawbackVestingAccount(ctx.Context(), createClawbackMsg)
	if err != nil {
		return err
	}

	fundAccMsg := &types.MsgFundVestingAccount{
		FunderAddress:  Account1,
		VestingAddress: Account1,
		StartTime:      time.Unix(VestingStartTime_12z83xmr, 0),
		// TODO: add vesting and lockup periods
	}

	// Then, fund the account
	err = FundVestingAccount(ctx, evk, sk, ak, bk, fundAccMsg)
	if err != nil {
		return err
	}

	return nil
}

func MigrateAccount2(ctx sdk.Context, ak authkeeper.AccountKeeper) error {
	// Get account
	account := ak.GetAccount(ctx, sdk.AccAddress(Account2))
	if account == nil {
		return nil
	}
	// change the start_time to Sept 4, 2022. The ugprade goes live on or after Sept 4, 2023, so the first year vest is still enforced
	// (the account was previously set to start on Sept 4, 2023)
	account.(*vesting.ContinuousVestingAccount).StartTime = VestingStartTime_1nwyvkxm
	// NOTE: we shouldn't have to update delegated_vesting on the BaseAccount. That's because,
	// DF (delegated free) and DV (delegated vesting) coins are set on (un)delegation and are point-in-time.
	// So, delegated_vesting overcounts how many tokens are vesting. Whenever an undelegation occurs, DF and DV should be set correctly.
	// See: https://github.com/cosmos/cosmos-sdk/commit/c5238b0d1ecfef8be3ccdaee02d23ee93ef9c69b
	// set the account
	ak.SetAccount(ctx, account)
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
				// sdk.NewAttribute(sdk.AttributeKeySender, msg.FunderAddress),
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
