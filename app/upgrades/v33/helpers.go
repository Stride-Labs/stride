package v33

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/migrations/v4/legacy"
	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
)

// SnapshotValidatorsFromICS reads the current CCV validator set from the
// consumer keeper and converts it into a slice of POA Validators ready to
// be passed to poaKeeper.InitGenesis.
//
// Each consensus address is joined to a moniker via the embedded
// ValidatorMonikers map (sourced from validators.json), and that moniker is
// joined to a Stride-side operator address via v33.FrozenValidatorSet — the
// same address the existing reward-allocation pipeline pays out to. Both joins
// must succeed; a missing entry on either side is a configuration drift
// between the two sources of truth and halts the upgrade.
func SnapshotValidatorsFromICS(
	ctx sdk.Context,
	consumerKeeper ccvconsumerkeeper.Keeper,
) ([]poatypes.Validator, error) {
	ccVals := consumerKeeper.GetAllCCValidator(ctx)
	if len(ccVals) > ExpectedValidatorCount {
		return nil, fmt.Errorf(
			"expected at most %d validators in consumer keeper, got %d",
			ExpectedValidatorCount, len(ccVals),
		)
	}
	if len(ccVals) < ExpectedValidatorCount {
		ctx.Logger().Error(fmt.Sprintf(
			"v33: expected %d validators in consumer keeper, got %d — "+
				"a validator may have been jailed; proceeding with current set",
			ExpectedValidatorCount, len(ccVals),
		))
	}

	operatorByMoniker := make(map[string]string, len(FrozenValidatorSet))
	for _, v := range FrozenValidatorSet {
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
				"validator %s (moniker %q) has no entry in v33.FrozenValidatorSet",
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

// PruneMalformedInFlightPackets deletes packet-forward-middleware in-flight packet records whose
// stored timeout height cannot be parsed. It must run before mm.RunMigrations.
//
// ibc-go's PFM consensus version 3→4 migration iterates the entire packetforward store with no
// prefix and rebuilds each record's packet via InFlightPacket.ChannelPacket(), which parses the
// stored height with clienttypes.MustParseHeight. That variant panics instead of returning an
// error, so one unparseable record aborts the upgrade on every node at the same height - a chain
// halt needing an emergency binary, not a failed upgrade.
//
// Stride mainnet carries exactly one such record: key 0x00, no refund routing, no packet data,
// and a timeout height of "" where the 25 legitimate entries all hold "0-0". It is not a real
// forward and has nothing to refund. We match on "height does not parse" rather than on that key
// so the guard still holds if another stray entry appears before the upgrade height.
//
// A record that fails to parse but *does* carry refund routing or packet data is deliberately not
// deleted - dropping one would silently abandon a user's refund. Those halt the upgrade for a
// human to triage instead.
func PruneMalformedInFlightPackets(
	ctx sdk.Context,
	cdc codec.Codec,
	pfmStoreKey *storetypes.KVStoreKey,
) (int, error) {
	store := ctx.KVStore(pfmStoreKey)

	malformedKeys, err := collectMalformedInFlightPacketKeys(ctx, cdc, store)
	if err != nil {
		return 0, err
	}

	// Deleted after the iterator is closed - mutating the store underneath an open iterator is
	// not defined behaviour.
	for _, key := range malformedKeys {
		store.Delete(key)
		ctx.Logger().Info(fmt.Sprintf("v33: pruned malformed in-flight packet at key %X", key))
	}

	return len(malformedKeys), nil
}

// collectMalformedInFlightPacketKeys returns the store keys of in-flight packets that would panic
// the PFM migration. Records it cannot unmarshal are left alone: the migration reports those as a
// clean error rather than a panic, so there is no need to risk deleting bytes we cannot interpret.
func collectMalformedInFlightPacketKeys(
	ctx sdk.Context,
	cdc codec.Codec,
	store storetypes.KVStore,
) ([][]byte, error) {
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	malformedKeys := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		var packet legacy.InFlightPacket
		if err := cdc.Unmarshal(iterator.Value(), &packet); err != nil {
			ctx.Logger().Error(fmt.Sprintf(
				"v33: in-flight packet at key %X does not unmarshal, leaving for the migration to report: %v",
				iterator.Key(), err,
			))
			continue
		}

		if _, err := clienttypes.ParseHeight(packet.PacketTimeoutHeight); err == nil {
			continue
		}

		if isRefundable(packet) {
			return nil, fmt.Errorf(
				"in-flight packet at key %X has unparseable timeout height %q but carries refund "+
					"routing (%s/%s seq %d); refusing to drop a refundable packet",
				iterator.Key(), packet.PacketTimeoutHeight,
				packet.RefundPortId, packet.RefundChannelId, packet.RefundSequence,
			)
		}

		// iterator.Key() is only valid for this iteration, so keep a copy.
		malformedKeys = append(malformedKeys, append([]byte(nil), iterator.Key()...))
	}

	return malformedKeys, nil
}

// isRefundable reports whether an in-flight packet carries enough routing to actually refund
// anyone. The degenerate mainnet record has none of it, so deleting it forfeits nothing.
func isRefundable(packet legacy.InFlightPacket) bool {
	return packet.RefundChannelId != "" ||
		packet.RefundPortId != "" ||
		packet.PacketSrcChannelId != "" ||
		packet.PacketSrcPortId != "" ||
		packet.RefundSequence != 0 ||
		len(packet.PacketData) != 0
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
