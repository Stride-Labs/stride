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
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalUndelegateCallbackArgs | %s", err.Error()))
		return unmarshalledRedemptionCallback, err
	}
	return unmarshalledRedemptionCallback, nil
}

func RedemptionCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement_Result, args []byte) error {
	logMsg := fmt.Sprintf("RedemptionCallback executing packet: %d, source: %s %s, dest: %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
	k.Logger(ctx).Info(logMsg)

	if ack == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "ack is nil")
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

	for _, epochNumber := range redemptionCallback.UnbondingEpochNumbers {
		epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
		if !found {
			errMsg := fmt.Sprintf("Epoch unbonding record not found for epoch #%d", epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}

		// TODO: Update with nondeterministic loop
		hostZoneUnbonding := epochUnbondingRecord.HostZoneUnbondings[hostZoneId]
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_TRANSFERRED

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
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}
	return nil
}
