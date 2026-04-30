package v33

import (
	"encoding/hex"
	"fmt"

	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	"github.com/Stride-Labs/stride/v32/utils"
)

// SnapshotValidatorsFromICS reads the current CCV validator set from the
// consumer keeper and converts it into a slice of POA Validators ready to
// be passed to poaKeeper.InitGenesis.
//
// Each consensus address is joined to a moniker via the embedded
// ValidatorMonikers map (sourced from validators.json), and that moniker is
// joined to a Stride-side operator address via utils.PoaValidatorSet — the
// same address the existing reward-allocation pipeline pays out to. Both joins
// must succeed; a missing entry on either side is a configuration drift
// between the two sources of truth and halts the upgrade.
func SnapshotValidatorsFromICS(
	ctx sdk.Context,
	consumerKeeper ccvconsumerkeeper.Keeper,
) ([]poatypes.Validator, error) {
	ccVals := consumerKeeper.GetAllCCValidator(ctx)
	if len(ccVals) != ExpectedValidatorCount {
		return nil, fmt.Errorf(
			"expected %d validators in consumer keeper, got %d",
			ExpectedValidatorCount, len(ccVals),
		)
	}

	operatorByMoniker := make(map[string]string, len(utils.PoaValidatorSet))
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}

	poaVals := make([]poatypes.Validator, 0, len(ccVals))
	for _, ccVal := range ccVals {
		consPubKey, err := ccVal.ConsPubKey()
		if err != nil {
			return nil, errorsmod.Wrapf(err,
				"failed to decode cons pubkey for validator %x", ccVal.Address)
		}
		pubKeyAny, err := codectypes.NewAnyWithValue(consPubKey)
		if err != nil {
			return nil, errorsmod.Wrapf(err,
				"failed to wrap cons pubkey for validator %x", ccVal.Address)
		}

		hexAddr := hex.EncodeToString(ccVal.Address)
		moniker, ok := ValidatorMonikers[hexAddr]
		if !ok {
			return nil, fmt.Errorf(
				"validator %s has no moniker in v33 validators.json", hexAddr,
			)
		}
		operatorAddr, ok := operatorByMoniker[moniker]
		if !ok {
			return nil, fmt.Errorf(
				"validator %s (moniker %q) has no entry in utils.PoaValidatorSet",
				hexAddr, moniker,
			)
		}

		poaVals = append(poaVals, poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  ccVal.Power,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         moniker,
				OperatorAddress: operatorAddr,
			},
		})
	}

	return poaVals, nil
}

// InitializePOA seeds POA's KV store with the given validator set and admin.
// Mirrors the canonical SDK sample at
// cosmos-sdk/enterprise/poa/examples/migrate-from-pos/sample_upgrades/upgrade_handler.go.
//
// WithBlockHeight(0) is required: POA's CreateValidator path calls
// GetTotalPower, which only treats "no total power yet" as a non-error
// case when ctx.BlockHeight() == 0 (enterprise/poa/x/poa/keeper/validator.go).
//
// The keeper-level InitGenesis returns ([]abci.ValidatorUpdate, error); we
// discard the updates because an upgrade handler returns a VersionMap, not
// ABCI updates. The next EndBlock will reap and emit anything still queued.
func InitializePOA(
	ctx sdk.Context,
	cdc codec.Codec,
	poaKeeper *poakeeper.Keeper,
	adminAddress string,
	validators []poatypes.Validator,
) error {
	if _, err := sdk.AccAddressFromBech32(adminAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid admin address: %s", adminAddress)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx).WithBlockHeight(0)

	genesis := &poatypes.GenesisState{
		Params:     poatypes.Params{Admin: adminAddress},
		Validators: validators,
		// AllocatedFees intentionally omitted — fresh POA init has no
		// pre-existing per-validator fee allocations to restore.
	}

	_, err := poaKeeper.InitGenesis(sdkCtx, cdc, genesis)
	return err
}

// SweepICSModuleAccounts moves any residual balance from the two ICS-era
// reward module accounts (cons_redistribute and cons_to_send_to_provider)
// to the community pool. After v33, no module deposits to these accounts;
// any leftover balance would be permanently stranded otherwise.
func SweepICSModuleAccounts(
	ctx sdk.Context,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
) error {
	accountsToSweep := []string{
		ccvconsumertypes.ConsumerRedistributeName,
		ccvconsumertypes.ConsumerToSendToProviderName,
	}

	for _, moduleName := range accountsToSweep {
		moduleAddr := accountKeeper.GetModuleAddress(moduleName)
		balance := bankKeeper.GetAllBalances(ctx, moduleAddr)
		if balance.IsZero() {
			ctx.Logger().Info(fmt.Sprintf("v33: %s is empty, skipping sweep", moduleName))
			continue
		}

		if err := distrKeeper.FundCommunityPool(ctx, balance, moduleAddr); err != nil {
			return errorsmod.Wrapf(err, "failed to fund community pool from %s", moduleName)
		}
		ctx.Logger().Info(fmt.Sprintf("v33: swept %s from %s to community pool", balance, moduleName))
	}
	return nil
}
