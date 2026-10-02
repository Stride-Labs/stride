package v35

import (
	"context"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v35/x/autopilot/keeper"
	icaoraclekeeper "github.com/Stride-Labs/stride/v35/x/icaoracle/keeper"
	icqkeeper "github.com/Stride-Labs/stride/v35/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v35/x/stakeibc/keeper"
)

// CreateUpgradeHandler returns the v35 upgrade handler, the wind-down upgrade
// (docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md §5) extended with the
// authority hand-off and mass undelegation
// (docs/superpowers/specs/2026-10-02-v35-authority-and-undelegation-design.md §3, "authority
// spec"). Every step logs and skips on missing state, except the writes that would leave the
// chain without an upgrade path if skipped: the wasm upload-access write and the consensus
// authority, gov deposit and staking max-entries writes fail the upgrade, as does a failure to
// list the delegations to undelegate. The steps run in
// this order:
//  1. RunMigrations.
//  2. Turn off autopilot stakeibc and drop liquid stake / redeem stake from the ICA host allow-list.
//  3. Restrict wasm code upload to gov, then move the deploy key's contract admins to gov.
//  4. Mark comdex-1 deprecated and delete the dYdX trade route.
//  5. Deactivate the ICA oracles and empty the rate limiter.
//  6. Reset stale DelegationChangesInProgress flags on zones with no ICA in flight.
//  7. Purge haqq's pending slash-path ICQs, then all withdrawal-balance and calibration ICQs.
//  8. Apply the haqq delegation delta table (after its ICQs are gone).
//  9. Requeue the one failed LSM detokenization with its amount reduced by one.
//  10. Set the consensus-params authority to the team multisig (authority spec §3).
//  11. Close gov submission by raising both deposits above total supply.
//  12. Raise staking max unbonding entries to 100.
//  13. Drop delegate, redelegate, create validator and cancel unbonding from the ICA host allow-list.
//  14. Undelegate every delegation in full, skipping and logging any single one that fails.
//
// icaHostKeeper and ratelimitKeeper are pointers because their methods have pointer
// receivers. The ICA controller and channel keepers used by the stale-flag reset are read
// through the stakeibc keeper's exported fields, as is the records keeper used by the LSM reset.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
	autopilotKeeper autopilotkeeper.Keeper,
	icaHostKeeper *icahostkeeper.Keeper,
	wasmKeeper wasmkeeper.Keeper,
	ratelimitKeeper *ratelimitkeeper.Keeper,
	icaOracleKeeper icaoraclekeeper.Keeper,
	consensusParamsKeeper consensusparamkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (protocol wind-down)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Entry points that bypass the msg service router (spec §5)
		DisableAutopilotStakeibc(ctx, autopilotKeeper)
		RemoveStakeibcFromICAHostAllowList(ctx, icaHostKeeper)

		// Wasm control to gov: the upload-access write is the one step that may fail the upgrade
		if err := SetWasmUploadAccessToGov(ctx, wasmKeeper); err != nil {
			return vm, err
		}
		MoveDeployKeyContractAdminsToGov(ctx, wasmKeeper)

		// Stakeibc state flips (spec §5)
		DeprecateComdex(ctx, stakeibcKeeper)
		DeleteDydxTradeRoute(ctx, stakeibcKeeper)

		// Oracles and rate limits (spec §5)
		DeactivateICAOracles(ctx, icaOracleKeeper)
		RemoveAllRateLimits(ctx, ratelimitKeeper)

		// Stale DelegationChangesInProgress flags on zones with nothing in flight (spec §5)
		ResetStaleDelegationChangesInProgress(ctx, stakeibcKeeper)

		// Pending ICQs (spec §5): the haqq slash-path purge runs before the haqq delta table
		PurgeHaqqSlashQueries(ctx, icqKeeper, stakeibcKeeper)
		PurgeWithdrawalBalanceQueries(ctx, icqKeeper)
		PurgeCalibrationQueries(ctx, icqKeeper)

		// Haqq delegation reconciliation, after its slash-path ICQs are gone (spec §5)
		ReconcileHaqqDelegations(ctx, stakeibcKeeper)

		// The failed LSM detokenization, retried by the EndBlocker with one token less (spec §5)
		ResetFailedLSMDeposit(ctx, stakeibcKeeper.RecordsKeeper)

		// Upgrade path without gov (authority spec §3): these writes fail the upgrade if they fail,
		// because without them the chain has no way to upgrade once stake is gone
		if err := SetConsensusAuthority(ctx, consensusParamsKeeper); err != nil {
			return vm, err
		}
		if err := CloseGovSubmission(ctx, govKeeper); err != nil {
			return vm, err
		}
		if err := RaiseMaxUnbondingEntries(ctx, stakingKeeper); err != nil {
			return vm, err
		}

		// Close the ICA path that could re-lock STRD, then unbond everything (authority spec §3)
		RemoveStakingFromICAHostAllowList(ctx, icaHostKeeper)
		if err := UndelegateAllDelegations(ctx, stakingKeeper); err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}
