package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func (k Keeper) MarshalRebalanceCallbackArgs(ctx sdk.Context, rebalanceCallback types.RebalanceCallback) ([]byte, error) {
	out, err := proto.Marshal(&rebalanceCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRebalanceCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalRebalanceCallbackArgs(ctx sdk.Context, rebalanceCallback []byte) (*types.RebalanceCallback, error) {
	unmarshalledRebalanceCallback := types.RebalanceCallback{}
	if err := proto.Unmarshal(rebalanceCallback, &unmarshalledRebalanceCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRebalanceCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledRebalanceCallback, nil
}

func RebalanceCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	k.Logger(ctx).Info("RebalanceCallback executing", "packet", packet)
	if ack == nil {
		// timeout
		k.Logger(ctx).Error(fmt.Sprintf("RebalanceCallback timeout, ack is nil, packet %v", packet))
		return nil
	}

	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}

	if len(txMsgData.Data) == 0 {
		// failed transaction
		k.Logger(ctx).Error(fmt.Sprintf("RebalanceCallback tx failed, ack is empty (ack error), packet %v", packet))
		return nil
	}

	// deserialize the args
	rebalanceCallback, err := k.UnmarshalRebalanceCallbackArgs(ctx, args)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal rebalance callback args | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("RebalanceCallback %v", rebalanceCallback))
	hostZone := rebalanceCallback.GetHostZoneId()
	zone, found := k.GetHostZone(ctx, hostZone)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", hostZone)
	}

	// update the host zone
	rebalancings := rebalanceCallback.GetRebalancings()
	// assemble a map from validatorAddress -> validator
	valAddrMap := make(map[string]*types.Validator)
	for _, val := range zone.GetValidators() {
		valAddrMap[val.GetAddress()] = val
	}
	for _, rebalancing := range rebalancings {
		srcValidator := rebalancing.GetSrcValidator()
		dstValidator := rebalancing.GetDstValidator()
		amt := rebalancing.Amt
		if _, valFound := valAddrMap[srcValidator]; valFound {
			valAddrMap[srcValidator].DelegationAmt = valAddrMap[srcValidator].DelegationAmt.Sub(amt)
		} else {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator not found %s", srcValidator)
		}
		if _, valFound := valAddrMap[dstValidator]; valFound {
			valAddrMap[dstValidator].DelegationAmt = valAddrMap[dstValidator].DelegationAmt.Add(amt)
		} else {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator not found %s", dstValidator)
		}
	}
	k.SetHostZone(ctx, zone)

	return nil
}
