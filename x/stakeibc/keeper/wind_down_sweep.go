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

	"github.com/Stride-Labs/stride/v34/utils"
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

// SweepTokensOffStride sends every listed denom each listed holder owns to the holder's own
// address bytes on the destination chain (spec §7): Stride-native denoms to Osmosis, vouchers
// back one hop over the channel they arrived on. Destinations are resolved once, before any
// address is read, so a bad denom rejects the batch and a bad holder only skips itself.
// Returns how many transfers were submitted and how many addresses were skipped
func (k Keeper) SweepTokensOffStride(
	ctx sdk.Context,
	msg *types.MsgSweepTokensOffStride,
) (numTransfers uint64, numSkipped uint64, err error) {
	destinations := map[string]sweepDestination{}
	for _, denom := range msg.Denoms {
		destination, err := k.resolveSweepDestination(ctx, denom)
		if err != nil {
			return 0, 0, err
		}
		destinations[denom] = destination
	}

	escrows := k.transferEscrowAddresses(ctx)
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())

	for _, holderBech32 := range msg.Addresses {
		holder := sdk.MustAccAddressFromBech32(holderBech32) // validated in ValidateBasic
		if reason, skip := k.sweepSkipReason(ctx, holder, escrows); skip {
			emitSweepSkippedEvent(ctx, holderBech32, reason)
			numSkipped++
			continue
		}

		for _, denom := range msg.Denoms {
			// Spendable, not total: the ICS-20 escrow is a bank send, which refuses coins a vesting
			// schedule still locks. Sweeping the total would fail the whole batch for every vesting
			// account with locked STRD, so the locked remainder stays and only what can move moves
			balance := k.bankKeeper.SpendableCoin(ctx, holder, denom)
			if balance.IsZero() {
				continue
			}

			destination := destinations[denom]
			receiver := sdk.MustBech32ifyAddressBytes(destination.Bech32Prefix, holder)
			transfer := transfertypes.MsgTransfer{
				SourcePort:       transfertypes.PortID,
				SourceChannel:    destination.ChannelId,
				Token:            balance,
				Sender:           holderBech32,
				Receiver:         receiver,
				TimeoutTimestamp: timeoutTimestamp,
				Memo:             "",
			}
			// A failed submission (closed channel, send disabled) is a batch problem, not a
			// holder problem: reject the whole tx so ops fix the cause and resubmit
			if _, err := k.RecordsKeeper.TransferKeeper.Transfer(ctx, &transfer); err != nil {
				return 0, 0, errorsmod.Wrapf(err, "unable to sweep %s from %s over %s",
					balance.String(), holderBech32, destination.ChannelId)
			}

			emitSweepTransferEvent(ctx, holderBech32, balance, destination.ChannelId, receiver)
			numTransfers++
		}
	}

	k.Logger(ctx).Info(fmt.Sprintf("Sweep submitted %d transfers for %d denoms across %d addresses (%d skipped)",
		numTransfers, len(msg.Denoms), len(msg.Addresses), numSkipped))
	return numTransfers, numSkipped, nil
}

func emitSweepSkippedEvent(ctx sdk.Context, address, reason string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSweepSkipped,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeySweepAddress, address),
			sdk.NewAttribute(types.AttributeKeySweepReason, reason),
		),
	)
}

func emitSweepTransferEvent(ctx sdk.Context, address string, amount sdk.Coin, channelId, receiver string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSweepTransfer,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeySweepAddress, address),
			sdk.NewAttribute(types.AttributeKeySweepDenom, amount.Denom),
			sdk.NewAttribute(types.AttributeKeySweepAmount, amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeySweepChannel, channelId),
			sdk.NewAttribute(types.AttributeKeySweepReceiver, receiver),
		),
	)
}
