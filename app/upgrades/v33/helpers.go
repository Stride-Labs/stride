package v33

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"

	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
)

// SnapshotValidatorsFromICS reads the current CCV validator set from the
// consumer keeper and converts it into a slice of POA Validators ready to
// be passed to poaKeeper.InitGenesis.
//
// Monikers are looked up from the pre-baked ValidatorMonikers map (keyed by
// hex of the consensus address). Missing entries fall back to empty string.
//
// ICS validators have no Stride-side operator address (the operator key lives
// on the Hub). POA's CreateValidator requires a valid sdk.AccAddress for
// OperatorAddress that (a) is unique per validator and (b) is distinct from the
// validator's consensus pubkey address. We synthesize one by hashing
// "poa-operator:" + consensusAddress bytes, which is deterministic and can be
// updated later via MsgUpdateValidators once real operator keys are known.
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

		operatorAddr := syntheticOperatorAddress(ccVal.Address)

		moniker := ValidatorMonikers[hex.EncodeToString(ccVal.Address)]

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

// syntheticOperatorAddress derives a unique, deterministic sdk.AccAddress
// placeholder for a validator whose real operator key is unknown. It hashes
// "poa-operator:" + consensusAddr and takes the first 20 bytes, guaranteeing
// uniqueness across validators while remaining distinct from the consensus key.
func syntheticOperatorAddress(consensusAddr []byte) string {
	hash := sha256.Sum256(append([]byte("poa-operator:"), consensusAddr...))
	return sdk.AccAddress(hash[:20]).String()
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
