package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func (k Keeper) MarshalRedemptionCallbackArgs(ctx sdk.Context, redemptionCallback types.RedemptionCallback) ([]byte, error) {
	out, err := proto.Marshal(&redemptionCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRedemptionCallbackArgs | %s", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalRedemptionCallbackArgs(ctx sdk.Context, redemptionCallback []byte) (types.RedemptionCallback, error) {
	unmarshalledRedemptionCallback := types.RedemptionCallback{}
	if err := proto.Unmarshal(redemptionCallback, &unmarshalledRedemptionCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRedemptionCallbackArgs | %s", err.Error()))
		return unmarshalledRedemptionCallback, err
	}
	return unmarshalledRedemptionCallback, nil
}

func RedemptionCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	logMsg := fmt.Sprintf("RedemptionCallback executing packet: %d, source: %s %s, dest: %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
	k.Logger(ctx).Info(logMsg)

	// unmarshal the callback args and get the host zone
	redemptionCallback, err := k.UnmarshalRedemptionCallbackArgs(ctx, args)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal redemption callback args | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("RedemptionCallback, HostZone: %s", redemptionCallback.HostZoneId))
	hostZoneId := redemptionCallback.HostZoneId
	zone, found := k.GetHostZone(ctx, hostZoneId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", hostZoneId)
	}

	if ack == nil {
		// handle timeout
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback timeout, ack is nil, packet %v", packet))
		return nil
	}

	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback txMsgData could not be parsed, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}

	if len(txMsgData.Data) == 0 {
		// handle tx failure
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback tx failed, txMsgData is empty, ack error, packet %v", packet))
		err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, zone.ChainId, redemptionCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
		if err != nil {
			return err
		}
		return nil
	}

	err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, zone.ChainId, redemptionCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_CLAIMABLE)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION] completed on %s", hostZoneId))
	return nil
}
