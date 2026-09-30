// x/stakeibc/keeper/wind_down_transfer_from_ica.go
package keeper

import (
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	if types.OsmosisVaultAddress == "" {
		return types.ErrOsmosisVaultNotConfigured
	}
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	icaMsg, err := BuildTransferFromIcaMsg(hostZone, msg.IcaType, msg.Amount, timeoutTimestamp)
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

// BuildTransferFromIcaMsg builds the message the ICA executes on the host: an ICS-20
// MsgTransfer over the zone's mapped channel to osmosis-1, or a bank MsgSend when the zone is
// osmosis-1 (mapped to an empty channel). It is exported so tests can assert every field.
func BuildTransferFromIcaMsg(
	hostZone types.HostZone,
	icaType types.ICAAccountType,
	amount sdk.Coin,
	timeoutTimestamp uint64,
) (proto.Message, error) {
	channelId, found := types.HostToOsmosisTransferChannel[hostZone.ChainId]
	if !found {
		return nil, types.ErrNoOsmosisChannelForHostZone.Wrapf("no channel to osmosis configured for %s", hostZone.ChainId)
	}
	icaAddress, err := windDownIcaAddress(hostZone, icaType)
	if err != nil {
		return nil, err
	}

	if channelId == "" {
		return &banktypes.MsgSend{
			FromAddress: icaAddress,
			ToAddress:   types.OsmosisVaultAddress,
			Amount:      sdk.NewCoins(amount),
		}, nil
	}
	return &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    channelId,
		Token:            amount,
		Sender:           icaAddress,
		Receiver:         types.OsmosisVaultAddress,
		TimeoutTimestamp: timeoutTimestamp,
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
