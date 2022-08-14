package keeper

import (
	"fmt"

	"github.com/spf13/cast"

	"github.com/golang/protobuf/proto"

	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// ___________________________________________________________________________________________________

// ICACallbacks wrapper struct for interchainstaking keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, *channeltypes.Acknowledgement_Result, []byte) error

type ICACallbacks struct {
	k            Keeper
	icacallbacks map[string]ICACallback
}

var _ icacallbackstypes.ICACallbackHandler = ICACallbacks{}

func (k Keeper) ICACallbackHandler() ICACallbacks {
	return ICACallbacks{k, make(map[string]ICACallback)}
}

//callback handler
func (c ICACallbacks) CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ack *channeltypes.Acknowledgement_Result, args []byte) error {
	return c.icacallbacks[id](c.k, ctx, packet, ack, args)
}

func (c ICACallbacks) HasICACallback(id string) bool {
	_, found := c.icacallbacks[id]
	return found
}

func (c ICACallbacks) AddICACallback(id string, fn interface{}) icacallbackstypes.ICACallbackHandler {
	c.icacallbacks[id] = fn.(ICACallback)
	return c
}

func (c ICACallbacks) RegisterICACallbacks() icacallbackstypes.ICACallbackHandler {
	return c.
		AddICACallback("delegate", ICACallback(DelegateCallback)).
		AddICACallback("undelegate", ICACallback(UndelegateCallback)).
		AddICACallback("redemption", ICACallback(RedemptionCallback))
}

// ----------------------------------- Delegate Callback ----------------------------------- //
func (k Keeper) MarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback types.DelegateCallback) ([]byte, error) {
	out, err := proto.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback []byte) (*types.DelegateCallback, error) {
	unmarshalledDelegateCallback := types.DelegateCallback{}
	if err := proto.Unmarshal(delegateCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

func DelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement_Result, args []byte) error {
	k.Logger(ctx).Info("DelegateCallback executing", "packet", packet)

	if ack == nil {
		// transaction on the host chain failed
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "ack is nil")
	}

	// deserialize the args
	delegateCallback, err := k.UnmarshalDelegateCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("DelegateCallback %v", delegateCallback))
	hostZone := delegateCallback.GetHostZoneId()
	zone, found := k.GetHostZone(ctx, hostZone)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", hostZone)
	}
	recordId := delegateCallback.GetDepositRecordId()

	for _, splitDelegation := range delegateCallback.SplitDelegations {
		amount := cast.ToInt64(splitDelegation.Amount)
		validator := splitDelegation.Validator

		k.Logger(ctx).Info(fmt.Sprintf("incrementing stakedBal %d on %s", amount, validator))
		if amount < 0 {
			errMsg := fmt.Sprintf("Balance to stake was negative: %d", amount)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(sdkerrors.ErrLogic, errMsg)
		} else {
			zone.StakedBal += amount
			success := k.AddDelegationToValidator(ctx, zone, validator, amount)
			if !success {
				return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to add delegation to validator")
			}
			k.SetHostZone(ctx, zone)
		}
	}

	k.RecordsKeeper.RemoveDepositRecord(ctx, cast.ToUint64(recordId))
	return nil
}
