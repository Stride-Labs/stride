package keeper

import (
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"
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

func RedemptionCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, txMsgData *sdk.TxMsgData, args []byte) error {
	logMsg := fmt.Sprintf("RedemptionCallback on executing packet: %d, source: %s %s, dest: %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
	k.Logger(ctx).Info(logMsg)

	if txMsgData == nil {
		k.Logger(ctx).Error(fmt.Sprintf("RedemptionCallback timeout, txMsgData is nil, packet %v", packet))
		return nil
	} else if len(txMsgData.Data) == 0 {
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
	for _, epochNumber := range redemptionCallback.UnbondingEpochNumbers {
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

		// Mark the redemption records claimable
		userRedemptionRecords := hostZoneUnbonding.UserRedemptionRecords
		for _, recordId := range userRedemptionRecords {
			userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, recordId)
			if !found {
				k.Logger(ctx).Error("failed to find user redemption record")
				return sdkerrors.Wrapf(types.ErrRecordNotFound, "no user redemption record found for id (%s)", recordId)
			}
			if userRedemptionRecord.IsClaimable {
				k.Logger(ctx).Info("user redemption record is already claimable")
				continue
			}
			userRedemptionRecord.IsClaimable = true
			k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
		}
	}
	k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION] completed on %s", hostZoneId))
	return nil
}
