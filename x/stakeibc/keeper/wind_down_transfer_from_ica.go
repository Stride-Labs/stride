// x/stakeibc/keeper/wind_down_transfer_from_ica.go
package keeper

import (
	"time"

	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// TransferFromIca submits one ICA containing a transfer of `amount` from one of the zone's
// four funded ICAs to the Osmosis vault (spec §7). The receiver is the hard-coded vault and
// the channel is the zone's hard-coded host-side channel to osmosis-1; for osmosis-1 itself
// the ICA already lives on Osmosis, so the message is a bank send. There is no callback: a
// failed or timed-out transfer refunds the ICA on the host and ops resubmit.
func (k Keeper) TransferFromIca(ctx sdk.Context, msg *types.MsgTransferFromIca) error {
	if err := validateOsmosisVaultAddress(types.OsmosisVaultAddress); err != nil {
		return err
	}
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	icaMsg, err := BuildTransferFromIcaMsg(hostZone, msg.IcaType, msg.Amount, ctx.BlockTime())
	if err != nil {
		return err
	}

	owner := types.FormatHostZoneICAOwner(hostZone.ChainId, msg.IcaType)
	if err := k.SubmitICATxWithoutCallback(ctx, hostZone.ConnectionId, owner, []proto.Message{icaMsg}, timeoutTimestamp); err != nil {
		return errorsmod.Wrapf(err, "unable to submit %s ICA transfer for %s", msg.IcaType, msg.ChainId)
	}

	channelId := types.HostToOsmosisTransferChannel[hostZone.ChainId]
	k.Logger(ctx).Info(utils.LogWithHostZone(msg.ChainId,
		"Wind-down transfer of %v from the %s ICA to %s over %q", msg.Amount, msg.IcaType, types.OsmosisVaultAddress, channelId))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferFromIca,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyHostZone, msg.ChainId),
			sdk.NewAttribute(types.AttributeKeyIcaType, msg.IcaType.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyChannel, channelId),
			sdk.NewAttribute(types.AttributeKeyReceiver, types.OsmosisVaultAddress),
		),
	)
	return nil
}

// validateOsmosisVaultAddress checks the hard-coded receiver is a well-formed osmo address
// before any message is built, so a bad build-time value fails closed instead of losing funds
func validateOsmosisVaultAddress(vault string) error {
	if vault == "" {
		return types.ErrOsmosisVaultNotConfigured
	}
	prefix, decoded, err := bech32.DecodeAndConvert(vault)
	if err != nil {
		return types.ErrOsmosisVaultNotConfigured.Wrapf("vault address %q is not valid bech32: %s", vault, err)
	}
	if prefix != types.OsmosisBech32Prefix {
		return types.ErrOsmosisVaultNotConfigured.Wrapf("vault address prefix %q, expected %q", prefix, types.OsmosisBech32Prefix)
	}
	if len(decoded) != 20 && len(decoded) != 32 {
		return types.ErrOsmosisVaultNotConfigured.Wrapf("vault address decodes to %d bytes, expected 20 or 32", len(decoded))
	}
	return nil
}

// BuildTransferFromIcaMsg builds the message the ICA executes on the host: an ICS-20
// MsgTransfer over the zone's mapped channel to osmosis-1, or a bank MsgSend when the zone is
// osmosis-1 (mapped to an empty channel). It is exported so tests can assert every field.
// The inner transfer times out at twice the ICA packet's window (the caller gives the ICA packet
// one WindDownTransferTimeout): a late-relayed ICA packet then still leaves the host to Osmosis
// leg a full day.
func BuildTransferFromIcaMsg(
	hostZone types.HostZone,
	icaType types.ICAAccountType,
	amount sdk.Coin,
	blockTime time.Time,
) (proto.Message, error) {
	channelId, found := types.HostToOsmosisTransferChannel[hostZone.ChainId]
	if !found {
		return nil, types.ErrNoOsmosisChannelForHostZone.Wrapf("no channel to osmosis configured for %s", hostZone.ChainId)
	}
	icaAddress, err := windDownIcaAddress(hostZone, icaType)
	if err != nil {
		return nil, err
	}

	if hostZone.ChainId == types.OsmosisChainId {
		return &banktypes.MsgSend{
			FromAddress: icaAddress,
			ToAddress:   types.OsmosisVaultAddress,
			Amount:      sdk.NewCoins(amount),
		}, nil
	}
	// A non-osmosis zone must never fall through to a bank send
	if channelId == "" {
		return nil, types.ErrNoOsmosisChannelForHostZone.Wrapf("empty channel to osmosis configured for %s", hostZone.ChainId)
	}
	innerTimeout := utils.IntToUint(blockTime.Add(2 * types.WindDownTransferTimeout).UnixNano())
	return &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    channelId,
		Token:            amount,
		Sender:           icaAddress,
		Receiver:         types.OsmosisVaultAddress,
		TimeoutTimestamp: innerTimeout,
		Memo:             "",
	}, nil
}

// windDownIcaAddress resolves the ICA address for one of the four funded ICA types
func windDownIcaAddress(hostZone types.HostZone, icaType types.ICAAccountType) (string, error) {
	var address string
	switch icaType {
	case types.ICAAccountType_DELEGATION:
		address = hostZone.DelegationIcaAddress
	case types.ICAAccountType_WITHDRAWAL:
		address = hostZone.WithdrawalIcaAddress
	case types.ICAAccountType_FEE:
		address = hostZone.FeeIcaAddress
	case types.ICAAccountType_REDEMPTION:
		address = hostZone.RedemptionIcaAddress
	default:
		return "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ica type %s cannot be transferred from", icaType)
	}
	if address == "" {
		return "", types.ErrICAAccountNotFound.Wrapf("%s ICA for %s has no address", icaType, hostZone.ChainId)
	}
	return address, nil
}
