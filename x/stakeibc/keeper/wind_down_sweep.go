package keeper

import (
	"fmt"
	"strings"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// sweepDestination is where a denom leaves Stride: the transfer channel and the bech32 prefix
// the holder's own address bytes are encoded with on the other side (spec §7)
type sweepDestination struct {
	ChannelId    string
	Bech32Prefix string
}

// resolveSweepDestination decides a denom's destination once per tx. A Stride-native denom
// (every stToken, ustrd) goes to Osmosis. An ibc/ voucher goes back over the channel it
// arrived on (the outermost hop of its trace) so it unwinds exactly one hop, and only if that
// channel leads to a chain whose wallets derive the same address bytes as Stride
// (types.SweepUnwindChannels); anything else has no safe destination and rejects the batch
func (k Keeper) resolveSweepDestination(ctx sdk.Context, denom string) (sweepDestination, error) {
	ibcPrefix := transfertypes.DenomPrefix + "/"
	if !strings.HasPrefix(denom, ibcPrefix) {
		return sweepDestination{
			ChannelId:    types.StrideToOsmosisTransferChannelId,
			Bech32Prefix: types.OsmosisBech32Prefix,
		}, nil
	}

	hash, err := transfertypes.ParseHexHash(denom[len(ibcPrefix):])
	if err != nil {
		return sweepDestination{}, errorsmod.Wrapf(types.ErrSweepDestinationUnavailable, "invalid ibc denom %s: %s", denom, err)
	}
	trace, found := k.RecordsKeeper.TransferKeeper.GetDenom(ctx, hash)
	if !found || len(trace.Trace) == 0 {
		return sweepDestination{}, errorsmod.Wrapf(types.ErrSweepDestinationUnavailable, "no denom trace for %s", denom)
	}

	outerChannel := trace.Trace[0].ChannelId
	prefix, whitelisted := types.SweepUnwindChannels[outerChannel]
	if !whitelisted {
		return sweepDestination{}, errorsmod.Wrapf(types.ErrSweepDestinationUnavailable,
			"denom %s arrived over %s, which is not a whitelisted unwind channel", denom, outerChannel)
	}
	return sweepDestination{ChannelId: outerChannel, Bech32Prefix: prefix}, nil
}

// transferEscrowAddresses returns the escrow address of every transfer channel, keyed by
// bech32 string. Escrows hold the supply of every stToken that lives on another chain and are
// never swept
func (k Keeper) transferEscrowAddresses(ctx sdk.Context) map[string]bool {
	escrows := map[string]bool{}
	for _, channel := range k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, transfertypes.PortID) {
		escrows[transfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId).String()] = true
	}
	return escrows
}

// sweepSkipReason applies the per-address rules of spec §7: only a 20-byte address whose
// account is a plain or vesting account has a counterpart the same key controls on the
// destination chain. Everything else (escrows, module accounts, interchain accounts owned by
// other chains, 32-byte contract-style addresses, addresses with no account) is skipped, and
// the reason is what the event carries so the off-chain builder learns why it disagreed
func (k Keeper) sweepSkipReason(ctx sdk.Context, address sdk.AccAddress, escrows map[string]bool) (reason string, skip bool) {
	if len(address) != 20 {
		return "address is not 20 bytes", true
	}
	// Escrows are checked before the account lookup: an escrow that has never received a
	// transfer has no account yet, and it must still be named as an escrow, not "not found"
	if escrows[address.String()] {
		return "transfer escrow address", true
	}
	account := k.AccountKeeper.GetAccount(ctx, address)
	if account == nil {
		return "account not found", true
	}

	// The concrete types are listed on purpose: an interchain account embeds a BaseAccount, so
	// an interface check would admit it
	switch account.(type) {
	case *authtypes.BaseAccount,
		*vestingtypes.ContinuousVestingAccount,
		*vestingtypes.DelayedVestingAccount,
		*vestingtypes.PeriodicVestingAccount,
		*claimvestingtypes.StridePeriodicVestingAccount:
		return "", false
	case *icatypes.InterchainAccount:
		return fmt.Sprintf("account type %T is not sweepable", account), true
	default:
		return fmt.Sprintf("account type %T is not sweepable", account), true
	}
}
