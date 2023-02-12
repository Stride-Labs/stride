package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

type Metric struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ContractMsgUpdateMetric struct {
	UpdateMetric Metric `json:"update_metric"`
}

func (k msgServer) UpdateMetric(goCtx context.Context, msg *types.MsgUpdateMetric) (*types.MsgUpdateMetricResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, "JUNO")
	if !found {
		return &types.MsgUpdateMetricResponse{}, errors.New("Juno not found")
	}

	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return &types.MsgUpdateMetricResponse{}, errors.New("Delegation account not registered")
	}

	contractMsg := ContractMsgUpdateMetric{
		UpdateMetric: Metric{
			Key:   "key_from_stride",
			Value: "value_from_stride",
		},
	}

	contractMsgBz, err := json.Marshal(contractMsg)
	if err != nil {
		return &types.MsgUpdateMetricResponse{}, errors.New("Unable to unmarshal contract")
	}

	timeout := ctx.BlockTime().UnixNano() + time.Minute.Nanoseconds()
	msgs := []sdk.Msg{
		&types.MsgExecuteContract{
			Sender:   delegationAccount.Address,
			Contract: msg.ContractAddress,
			Msg:      contractMsgBz,
		},
	}
	fmt.Printf("Submitting UpdateMetric ICA with Timeout: %d\n", timeout)

	_, err = k.SubmitTxs(ctx, hostZone.ConnectionId, msgs, *hostZone.GetDelegationAccount(), uint64(timeout), "", []byte{})
	if err != nil {
		return &types.MsgUpdateMetricResponse{}, err
	}

	return &types.MsgUpdateMetricResponse{}, nil
}
