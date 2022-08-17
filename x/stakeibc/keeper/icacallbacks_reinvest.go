package keeper

import (
	"fmt"

	"github.com/spf13/cast"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"
)

func (k Keeper) MarshalReinvestCallbackArgs(ctx sdk.Context, reinvestCallback types.ReinvestCallback) ([]byte, error) {
	out, err := proto.Marshal(&reinvestCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalReinvestCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalReinvestCallbackArgs(ctx sdk.Context, reinvestCallback []byte) (*types.ReinvestCallback, error) {
	unmarshalledReinvestCallback := types.ReinvestCallback{}
	if err := proto.Unmarshal(reinvestCallback, &unmarshalledReinvestCallback); err != nil {
        k.Logger(ctx).Error(fmt.Sprintf("UnmarshalReinvestCallbackArgs %s", err.Error()))
		return nil, err
	}
	return &unmarshalledReinvestCallback, nil
}

func ReinvestCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement_Result, args []byte) error {
	k.Logger(ctx).Info("ReinvestCallback executing", "packet", packet)

	if ack == nil {
		// transaction on the host chain failed
		// don't create a record
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "ack is nil, packet %v", packet)
	}

	// deserialize the args
	reinvestCallback, err := k.UnmarshalReinvestCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	amount := reinvestCallback.ReinvestAmount.Amount
	denom := reinvestCallback.ReinvestAmount.Denom
	
	// fetch epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find epoch")
		return sdkerrors.Wrapf(types.ErrInvalidLengthEpochTracker, "no number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	epochNumber := strideEpochTracker.EpochNumber
	// create a new record so that rewards are reinvested
	amt, err := cast.ToInt64E(amount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to convert amount %v", err.Error()))
		return err
	}
	record := recordstypes.DepositRecord{
		Amount:             amt,
		Denom:              denom,
		HostZoneId:         reinvestCallback.HostZoneId,
		Status:             recordstypes.DepositRecord_STAKE,
		Source:             recordstypes.DepositRecord_WITHDRAWAL_ICA,
		DepositEpochNumber: epochNumber,
	}
	k.RecordsKeeper.AppendDepositRecord(ctx, record)

	// update the balance on the fee account TODO
	// store the chain id on the reinvestment callback
	// feeAccount := zone.GetFeeAccount()
	// update the balance on the fee account
	return nil
}
