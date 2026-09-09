package v34

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/Stride-Labs/stride/v33/utils"
)

// CreateUpgradeHandler returns the v34 upgrade handler, which swaps two POA
// validators. See docs/superpowers/specs/2026-09-09-v34-validator-swap-design.md.
//
// poaKeeper is a pointer because POA's keeper methods have pointer receivers.
// cdc unpacks the stored consensus-pubkey Anys when resolving the outgoing
// validators' consensus addresses.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	poaKeeper *poakeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (POA validator swap)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		if err := SwapPoaValidators(ctx, cdc, poaKeeper); err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}

type incomingValidator struct {
	consAddress sdk.ConsAddress
	validator   poatypes.Validator
}

type outgoingValidator struct {
	moniker     string
	consAddress sdk.ConsAddress
}

// SwapPoaValidators adds the incoming validators to the POA set and removes
// (power → 0) the outgoing ones.
//
// Consensus-safety invariant: each consensus address gets exactly ONE
// power-changing keeper call. POA queues one ABCI update per power change
// with no dedup, and CometBFT panics every node on a duplicate consensus
// address in a single block's update set. The 6 continuing validators are
// deliberately never touched.
func SwapPoaValidators(ctx sdk.Context, cdc codec.Codec, poaKeeper *poakeeper.Keeper) error {
	incoming, err := buildIncomingValidators()
	if err != nil {
		return err
	}
	outgoing, err := resolveOutgoingValidators(ctx, cdc, poaKeeper)
	if err != nil {
		return err
	}

	// Create before removing so total power never dips during the transition.
	// checkpoint=true so unallocated fees are checkpointed to the pre-existing
	// set before the new validator becomes eligible (genesis is the only
	// correct checkpoint=false caller).
	for _, validator := range incoming {
		ctx.Logger().Info(fmt.Sprintf("v34: adding POA validator %s", validator.validator.Metadata.Moniker))
		if err := poaKeeper.CreateValidator(ctx, validator.consAddress, validator.validator, true); err != nil {
			return fmt.Errorf("failed to create POA validator %s: %w", validator.validator.Metadata.Moniker, err)
		}
	}

	// Power 0 removes the validator from the active set; nil Metadata/PubKey
	// preserve the stored record so accrued fees stay withdrawable.
	for _, validator := range outgoing {
		ctx.Logger().Info(fmt.Sprintf("v34: removing POA validator %s", validator.moniker))
		if err := poaKeeper.UpdateValidator(ctx, validator.consAddress, poatypes.Validator{Power: 0}); err != nil {
			return fmt.Errorf("failed to remove POA validator %s: %w", validator.moniker, err)
		}
	}

	return nil
}

// buildIncomingValidators validates the incoming constants and joins each
// moniker to its payout address in utils.PoaValidatorSet.
//
// The pubkey is decoded into a concrete ed25519.PubKey (rather than accepting
// an arbitrary Any) because consensus params allow only ed25519 and
// keeper.CreateValidator does not validate key types — a wrong key type would
// reach CometBFT and halt the chain at EndBlock.
func buildIncomingValidators() ([]incomingValidator, error) {
	operatorByMoniker := make(map[string]string, len(utils.PoaValidatorSet))
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}

	incoming := make([]incomingValidator, 0, len(IncomingValidators))
	for _, entry := range IncomingValidators {
		if entry.ConsPubKeyBase64 == PlaceholderConsPubKey {
			return nil, fmt.Errorf("incoming validator %q still has a placeholder consensus pubkey", entry.Moniker)
		}
		keyBytes, err := base64.StdEncoding.DecodeString(entry.ConsPubKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("incoming validator %q consensus pubkey is not valid base64: %w", entry.Moniker, err)
		}
		if len(keyBytes) != ed25519.PubKeySize {
			return nil, fmt.Errorf("incoming validator %q consensus pubkey has %d bytes, expected %d (ed25519)",
				entry.Moniker, len(keyBytes), ed25519.PubKeySize)
		}

		operator, ok := operatorByMoniker[entry.Moniker]
		if !ok {
			return nil, fmt.Errorf("incoming validator %q has no entry in utils.PoaValidatorSet", entry.Moniker)
		}
		if utils.IsPlaceholderOperator(operator) {
			return nil, fmt.Errorf("incoming validator %q still has a placeholder payout address", entry.Moniker)
		}
		if _, err := sdk.AccAddressFromBech32(operator); err != nil {
			return nil, fmt.Errorf("incoming validator %q payout address is invalid: %w", entry.Moniker, err)
		}

		pubKey := &ed25519.PubKey{Key: keyBytes}
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to pack pubkey for %q: %w", entry.Moniker, err)
		}

		incoming = append(incoming, incomingValidator{
			consAddress: sdk.GetConsAddress(pubKey),
			validator: poatypes.Validator{
				PubKey: pubKeyAny,
				Power:  ValidatorPower,
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         entry.Moniker,
					OperatorAddress: operator,
				},
			},
		})
	}
	return incoming, nil
}

// resolveOutgoingValidators maps each outgoing moniker to exactly one
// consensus address in live POA state, erroring loudly on a missing or
// ambiguous match.
func resolveOutgoingValidators(ctx sdk.Context, cdc codec.Codec, poaKeeper *poakeeper.Keeper) ([]outgoingValidator, error) {
	validators, err := poaKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	consAddressesByMoniker := make(map[string][]sdk.ConsAddress)
	for _, validator := range validators {
		if validator.PubKey == nil {
			return nil, fmt.Errorf("POA validator has a nil consensus pubkey, cannot resolve its consensus address")
		}
		var pubKey cryptotypes.PubKey
		if err := cdc.UnpackAny(validator.PubKey, &pubKey); err != nil {
			return nil, fmt.Errorf("failed to unpack pubkey for POA validator: %w", err)
		}
		consAddress := sdk.GetConsAddress(pubKey)

		if validator.Metadata == nil {
			// No known code path leaves Metadata nil (this handler's own
			// removals pass nil to UpdateValidator, but that only skips
			// overwriting — the stored Metadata survives). Still, a
			// nil-Metadata record can never match an OutgoingMonikers entry
			// by moniker, so skip it defensively rather than panic — but log
			// so a later "outgoing validator not found" error isn't a
			// mysterious dead end.
			ctx.Logger().Warn(fmt.Sprintf("v34: POA validator with cons address %s has nil metadata, skipping", consAddress))
			continue
		}

		moniker := validator.Metadata.Moniker
		consAddressesByMoniker[moniker] = append(consAddressesByMoniker[moniker], consAddress)
	}

	outgoing := make([]outgoingValidator, 0, len(OutgoingMonikers))
	for _, moniker := range OutgoingMonikers {
		matches := consAddressesByMoniker[moniker]
		if len(matches) == 0 {
			return nil, fmt.Errorf("outgoing validator %q not found in POA state", moniker)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("outgoing validator moniker %q matches %d POA validators, refusing to guess", moniker, len(matches))
		}
		outgoing = append(outgoing, outgoingValidator{moniker: moniker, consAddress: matches[0]})
	}
	return outgoing, nil
}
