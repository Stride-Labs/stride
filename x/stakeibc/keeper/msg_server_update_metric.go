package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
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

type ContractMsgInstantiate struct {
	AdminAddress string `json:"admin_address"`
	ICAAddress   string `json:"ica_address"`
}

func (k Keeper) SetContractAddress(ctx sdk.Context, contractAddress string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix("ica-poc"))
	key := []byte("contract-address")
	store.Set(key, []byte(contractAddress))
}

func (k Keeper) GetContractAddress(ctx sdk.Context) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix("ica-poc"))
	key := []byte("contract-address")
	address := store.Get(key)
	if address == nil || len(address) == 0 {
		return ""
	}
	return string(address)
}

func (k msgServer) UpdateMetric(goCtx context.Context, msg *types.MsgUpdateMetric) (*types.MsgUpdateMetricResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var instantiateContract bool
	var contractAddress string
	contractAddress = k.GetContractAddress(ctx)
	if contractAddress == "" {
		instantiateContract = true
	} else {
		instantiateContract = false
	}

	hostZone, found := k.GetHostZone(ctx, "JUNO")
	if !found {
		return &types.MsgUpdateMetricResponse{}, errors.New("Juno not found")
	}

	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return &types.MsgUpdateMetricResponse{}, errors.New("Delegation account not registered")
	}

	var msgs []sdk.Msg
	if instantiateContract {
		fmt.Println("Instantiating contract")
		contractMsg := ContractMsgInstantiate{
			AdminAddress: delegationAccount.Address,
			ICAAddress:   delegationAccount.Address,
		}
		contractMsgBz, err := json.Marshal(contractMsg)
		if err != nil {
			return &types.MsgUpdateMetricResponse{}, errors.New("Unable to unmarshal contract")
		}
		msgs = []sdk.Msg{
			&types.MsgInstantiateContract{
				Sender: delegationAccount.Address,
				Admin:  delegationAccount.Address,
				CodeID: uint64(1),
				Label:  "ica-oracle",
				Msg:    contractMsgBz,
			},
		}
	} else {
		fmt.Println("Executing contract")
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
		msgs = []sdk.Msg{
			&types.MsgExecuteContract{
				Sender:   delegationAccount.Address,
				Contract: contractAddress,
				Msg:      contractMsgBz,
			},
		}
	}

	timeout := ctx.BlockTime().UnixNano() + time.Minute.Nanoseconds()

	fmt.Printf("Submitting ICA with Timeout: %d\n", timeout)

	_, err := k.SubmitTxs(ctx, hostZone.ConnectionId, msgs, *hostZone.GetDelegationAccount(), uint64(timeout), "", []byte{})
	if err != nil {
		return &types.MsgUpdateMetricResponse{}, err
	}

	return &types.MsgUpdateMetricResponse{}, nil
}
