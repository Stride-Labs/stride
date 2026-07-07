package v33

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"

	epochstypes "github.com/Stride-Labs/stride/v32/x/epochs/types"
	recordskeeper "github.com/Stride-Labs/stride/v32/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v32/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v32/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v32/x/stakeibc/types"
)

var UpgradeName = "v33"

const OsmosisChainId = "osmosis-1"

// ValidatorReconciliation records the true on-chain delegation for an Osmosis validator whose
// tracked Delegation drifted above reality after the v32 rebalance (an undelegation executed on
// Osmosis but its ack was lost, so the record was never decremented). See the remediation spec:
// docs/incidents/2026-v32-unbonding-remediation.md
type ValidatorReconciliation struct {
	Name              string
	Address           string
	OnChainDelegation sdkmath.Int // uosmo
}

// OsmosisPhantomDelegations lists the validators to reconcile and their current on-chain delegation.
//
// !!! RE-VERIFY these against live on-chain state immediately before the upgrade height !!!
// (query the delegation ICA osmo1npfl4vmmmf4yqhcemz95mvqujgdnlhrlxfzhlhz2gru8g2t749xqr9zm5e).
// cosmostation and chorus_one currently hold 0 on-chain; pryzmstakedrop holds ~60,662.8 OSMO.
var OsmosisPhantomDelegations = []ValidatorReconciliation{
	{Name: "cosmostation", Address: "osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4", OnChainDelegation: sdkmath.ZeroInt()},
	{Name: "chorus_one", Address: "osmovaloper15urq2dtp9qce4fyc85m6upwm9xul3049wh9czc", OnChainDelegation: sdkmath.ZeroInt()},
	{Name: "pryzmstakedrop", Address: "osmovaloper14n8pf9uxhuyxqnqryvjdr8g68na98wn5amq3e5", OnChainDelegation: sdkmath.NewInt(60_662_797_124)},
}

// CreateUpgradeHandler creates an SDK upgrade handler for v33
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	recordsKeeper recordskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Reconciling Osmosis phantom delegations...")
		if err := ReconcileOsmosisDelegations(ctx, stakeibcKeeper, recordsKeeper); err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

// ReconcileOsmosisDelegations fixes the OSMO delegation-record drift from the v32-rebalance incident
// without moving user funds:
//
//  1. For each affected validator, set its tracked Delegation to its true on-chain value and reduce
//     HostZone.TotalDelegations by the difference (the "phantom" that was never decremented).
//  2. Credit the reconciled total back as a DELEGATION_QUEUE deposit record, so it is counted in the
//     redemption rate (undelegatedBalance) and re-delegated from the delegation account's stranded
//     liquid by the normal delegation flow - the same mechanism reinvestment uses
//     (see x/stakeibc/keeper/icacallbacks_reinvest.go).
//
// Because the deposit-record credit equals the TotalDelegations reduction exactly, the redemption
// rate is unchanged. Once the phantom records are gone, the stuck OSMO redemptions retry against
// real validators and unbond/burn/pay through the normal path - no manual burn or sweep here.
func ReconcileOsmosisDelegations(ctx sdk.Context, sk stakeibckeeper.Keeper, rk recordskeeper.Keeper) error {
	hostZone, found := sk.GetHostZone(ctx, OsmosisChainId)
	if !found {
		// Skip rather than error: an error here fails the upgrade and halts the chain, and
		// non-mainnet environments (dockernet, testnets) don't have the osmosis-1 host zone
		ctx.Logger().Info(fmt.Sprintf("v33: host zone %s not found, skipping phantom delegation reconciliation", OsmosisChainId))
		return nil
	}

	totalReduction := sdkmath.ZeroInt()
	for _, reconciliation := range OsmosisPhantomDelegations {
		validator, index, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, reconciliation.Address)
		if !found {
			return errorsmod.Wrapf(stakeibctypes.ErrValidatorNotFound,
				"validator %s (%s) not found on %s", reconciliation.Name, reconciliation.Address, OsmosisChainId)
		}
		// Guard against a stale constant: we should only ever be REDUCING a phantom, never inflating.
		if validator.Delegation.IsNil() || validator.Delegation.LT(reconciliation.OnChainDelegation) {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"validator %s tracked delegation %v is below the expected on-chain amount %v; re-verify constants",
				reconciliation.Address, validator.Delegation, reconciliation.OnChainDelegation)
		}

		reduction := validator.Delegation.Sub(reconciliation.OnChainDelegation)
		validator.Delegation = reconciliation.OnChainDelegation
		hostZone.Validators[index] = &validator
		totalReduction = totalReduction.Add(reduction)

		ctx.Logger().Info(fmt.Sprintf("v33: reconciled %s delegation to %v (-%v)",
			reconciliation.Name, reconciliation.OnChainDelegation, reduction))
	}

	if totalReduction.IsZero() {
		ctx.Logger().Info("v33: no phantom delegation to reconcile on Osmosis")
		return nil
	}

	hostZone.TotalDelegations = hostZone.TotalDelegations.Sub(totalReduction)
	sk.SetHostZone(ctx, hostZone)

	// Credit the reconciled amount so the redemption rate is preserved and the stranded liquid is
	// re-delegated by the normal flow.
	strideEpoch, found := sk.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "stride epoch tracker not found")
	}
	rk.AppendDepositRecord(ctx, recordstypes.DepositRecord{
		Amount:             totalReduction,
		Denom:              hostZone.HostDenom,
		HostZoneId:         OsmosisChainId,
		Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
		Source:             recordstypes.DepositRecord_WITHDRAWAL_ICA,
		DepositEpochNumber: strideEpoch.EpochNumber,
	})

	ctx.Logger().Info(fmt.Sprintf("v33: created DELEGATION_QUEUE deposit record for %v %s (rate-preserving credit of reconciled phantom)",
		totalReduction, hostZone.HostDenom))
	return nil
}
