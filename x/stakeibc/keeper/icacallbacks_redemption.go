package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

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
	if ack == nil {
		// handle timeout
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback timeout, ack is nil, packet %v", packet))
		return nil
	}

	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback timeout, ack is nil, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}

	if len(txMsgData.Data) == 0 {
		// handle tx failure
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback tx failed, txMsgData is empty, ack error, packet %v", packet))
		return nil
	}

	// unmarshal the callback args and get the host zone
	redemptionCallback, err := k.UnmarshalRedemptionCallbackArgs(ctx, args)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal redemption callback args | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("RedemptionCallback, HostZone: %s", redemptionCallback.HostZoneId))

	hostZoneId := redemptionCallback.HostZoneId

	// Loop through all the epoch numbers that were stored with the callback (that identify the unbonding records)
	for _, epochNumber := range redemptionCallback.EpochUnbondingRecordIds {
		epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
		if !found {
			errMsg := fmt.Sprintf("Epoch unbonding record not found for epoch #%d", epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}

		// Update the unbonding status to TRANSFERRED
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZoneId)
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("Could not find host zone unbonding %d for host zone %s", epochUnbondingRecord.EpochNumber, hostZoneId))
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "Could not find host zone unbonding %d for host zone %s", epochUnbondingRecord.EpochNumber, hostZoneId)
		}
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_TRANSFERRED
		updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZoneId, hostZoneUnbonding)
		if !success {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v", epochUnbondingRecord.EpochNumber, hostZoneId, hostZoneUnbonding))
			return sdkerrors.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
		}
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)
	}
	k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION] completed on %s", hostZoneId))
	return nil
}
