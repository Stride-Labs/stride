// Package distrwrapper provides an in-tree replacement for the standard
// x/distribution AppModule's BeginBlock that:
//
//  1. Iterates BONDED staking validators by stake (not consensus voters),
//     avoiding the standard distribution module's `ValidatorByConsAddr` lookup.
//     Pre-v33 this lookup happened to succeed only for 5 of Stride's 8 ICS
//     consensus validators; the other 3 use ICS key-assignment with a
//     different consumer pubkey than their staking pubkey, which would cause
//     `distribution.AllocateTokens` to error and halt the chain.
//
//  2. Splits inflation 85/15 between staking validators and POA validators,
//     preserving the pre-v33 ConsumerRedistributionFraction on mainnet. The
//     15% slice goes into the POA module account, where POA's existing lazy
//     checkpoint (`checkpointAllValidators`) distributes it pro-rata across
//     POA validators on the next set change or `MsgWithdrawFees` call.
//
// The BeginBlock body is a near-verbatim copy of
// `interchain-security/v7/x/ccv/democracy/distribution/module.go`, with the
// fee source changed from `ConsumerRedistributeName` → `FeeCollectorName` and
// the 15% pre-routing to POA prepended. Everything else (genesis, services,
// queries) passes through to the embedded standard `x/distribution.AppModule`.
package distrwrapper

import (
	"context"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/exported"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// POAFraction is the share of inflation routed to POA validators, mirroring
// the pre-v33 ConsumerRedistributionFraction split (85% staking / 15% POA).
// Hardcoded here rather than parameterized: this matches mainnet's existing
// economic profile and any future change should go through a deliberate
// upgrade rather than runtime governance.
var POAFraction = math.LegacyMustNewDecFromStr("0.15")

// Interface conformance assertions. The module.* interfaces are deprecated in
// favor of appmodule.* extension interfaces, but the SDK still type-asserts
// against module.* in module.NewManager — same pattern used by ccvdistr and
// every other AppModule in this app, so we keep them and silence the lint.
//
//nolint:staticcheck // SA1019: see comment above
var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
)

// AppModuleBasic embeds the standard x/distribution AppModuleBasic.
type AppModuleBasic struct {
	distr.AppModuleBasic
}

// AppModule embeds the standard x/distribution AppModule and overrides
// BeginBlock with a stake-iteration allocation that mirrors the pre-v33
// ccvdistr behavior.
type AppModule struct {
	distr.AppModule

	keeper        keeper.Keeper
	accountKeeper distrtypes.AccountKeeper
	bankKeeper    distrtypes.BankKeeper
	stakingKeeper stakingkeeper.Keeper
}

// NewAppModule constructs the wrapper, embedding a standard distr.AppModule
// for everything other than BeginBlock.
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	ak distrtypes.AccountKeeper,
	bk distrtypes.BankKeeper,
	sk stakingkeeper.Keeper,
	subspace exported.Subspace,
) AppModule {
	return AppModule{
		AppModule:     distr.NewAppModule(cdc, keeper, ak, bk, sk, subspace),
		keeper:        keeper,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

// BeginBlock allocates the previous block's collected fees to staking +
// POA validators using the stake-iteration model (no consensus-address
// lookup). Mirrors ccvdistr's BeginBlock height-guard semantics.
func (am AppModule) BeginBlock(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	defer telemetry.ModuleMeasureSince(distrtypes.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker) //nolint:staticcheck // SA1019: ModuleMeasureSince deprecated, but ccvdistr uses it; switch with the SDK

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		return am.allocateTokens(ctx)
	}
	return nil
}

// allocateTokens routes fees from fee_collector:
//   - 15% → POA module account (POA's lazy checkpoint distributes to POA validators)
//   - 85% → distribution module → bonded staking validators by stake fraction
//   - rounding remainder + community-tax slice → community pool
func (am AppModule) allocateTokens(ctx sdk.Context) error {
	feeCollector := am.accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	feesCollectedInt := am.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	if feesCollectedInt.IsZero() {
		return nil
	}

	// Step 1: POA share. Truncate so we don't move more than is in fee_collector.
	feesDec := sdk.NewDecCoinsFromCoins(feesCollectedInt...)
	poaShareDec := feesDec.MulDecTruncate(POAFraction)
	poaShare, _ := poaShareDec.TruncateDecimal()

	if !poaShare.IsZero() {
		if err := am.bankKeeper.SendCoinsFromModuleToModule(
			ctx, authtypes.FeeCollectorName, poatypes.ModuleName, poaShare,
		); err != nil {
			return err
		}
	}

	// Step 2: remaining fees (≈85%) flow through the standard distribution
	// module to bonded staking validators. Below is a near-verbatim port of
	// ccvdistr.AllocateTokens with the fee source rewired.
	stakingShareInt := feesCollectedInt.Sub(poaShare...)
	if stakingShareInt.IsZero() {
		return nil
	}
	stakingShareDec := sdk.NewDecCoinsFromCoins(stakingShareInt...)

	if err := am.bankKeeper.SendCoinsFromModuleToModule(
		ctx, authtypes.FeeCollectorName, distrtypes.ModuleName, stakingShareInt,
	); err != nil {
		return err
	}

	feePool, err := am.keeper.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	vs := am.stakingKeeper.GetValidatorSet()
	totalBondedTokens := math.ZeroInt()
	if err := vs.IterateBondedValidatorsByPower(ctx, func(_ int64, validator stakingtypes.ValidatorI) bool {
		totalBondedTokens = totalBondedTokens.Add(validator.GetTokens())
		return false
	}); err != nil {
		return err
	}

	// No bonded validators → everything to community pool.
	if totalBondedTokens.IsZero() {
		feePool.CommunityPool = feePool.CommunityPool.Add(stakingShareDec...)
		return am.keeper.FeePool.Set(ctx, feePool)
	}

	communityTax, err := am.keeper.GetCommunityTax(ctx)
	if err != nil {
		return err
	}
	validatorFraction := math.LegacyOneDec().Sub(communityTax)

	remaining := stakingShareDec
	var allocErr error
	if err := vs.IterateBondedValidatorsByPower(ctx, func(_ int64, validator stakingtypes.ValidatorI) bool {
		if allocErr != nil {
			return true
		}
		powerFraction := math.LegacyNewDecFromInt(validator.GetTokens()).
			QuoTruncate(math.LegacyNewDecFromInt(totalBondedTokens))
		reward := stakingShareDec.MulDecTruncate(validatorFraction).MulDecTruncate(powerFraction)
		if err := am.keeper.AllocateTokensToValidator(ctx, validator, reward); err != nil {
			allocErr = err
			return true
		}
		remaining = remaining.Sub(reward)
		return false
	}); err != nil {
		return err
	}
	if allocErr != nil {
		return allocErr
	}

	// Truncation remainder + community-tax slice → community pool.
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	return am.keeper.FeePool.Set(ctx, feePool)
}
