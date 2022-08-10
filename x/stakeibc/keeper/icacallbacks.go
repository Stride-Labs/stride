package keeper

import (
	"fmt"

	"encoding/json"

	"github.com/spf13/cast"

	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"
)

// ___________________________________________________________________________________________________

// ICACallbacks wrapper struct for interchainstaking keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, []byte, []byte) error

type ICACallbacks struct {
	k         Keeper
	icacallbacks map[string]ICACallback
}

var _ icacallbackstypes.ICACallbackHandler = ICACallbacks{}

func (k Keeper) ICACallbackHandler() ICACallbacks {
	return ICACallbacks{k, make(map[string]ICACallback)}
}

//callback handler
func (c ICACallbacks) CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ack []byte, args []byte) error {
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
	a := c.
			AddICACallback("delegate", ICACallback(DelegateCallback)).
			AddICACallback("redemption", ICACallback(RedemptionCallback))
	return a.(ICACallbacks)
}

// -----------------------------------
// ICACallback Handlers
// -----------------------------------

// ----------------------------------- Delegate Callback ----------------------------------- //
func (k Keeper) MarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback types.DelegateCallback) []byte {
	out, err := proto.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
	}
	return out
}

func (k Keeper) UnmarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback []byte) types.DelegateCallback {
	unmarshalledDelegateCallback := types.DelegateCallback{}
	if err := proto.Unmarshal(delegateCallback, &unmarshalledDelegateCallback); err != nil {
        k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
	}
	return unmarshalledDelegateCallback
}

// QUESTION: Would it be cleaner to pass in ack as a bool (success / failure) here?
func DelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack []byte, args []byte) error {
	k.Logger(ctx).Info("DelegateCallback executing", "packet", packet, "ack", ack, "args", args)
	// deserialize the ack
	txMsgData, err := k.GetTxMsgData(ctx, ack)
	if err != nil {
		// ack failed, handle here
		return err
	}
	// do we need txMsgData?
	_ = txMsgData

	// deserialize the args
	delegateCallback := k.UnmarshalDelegateCallbackArgs(ctx, args)
	k.Logger(ctx).Info(fmt.Sprintf("DelegateCallback %v", delegateCallback))
	hostZone := delegateCallback.GetHostZoneId()
	zone, found := k.GetHostZone(ctx, hostZone)
	_ = zone
	if !found {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "host zone not found")
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

// ----------------------------------- redemption callback ----------------------------------- //
func (k Keeper) MarshalRedemptionCallbackArgs(ctx sdk.Context, redemptionCallback types.RedemptionCallback) []byte {
	out, err := proto.Marshal(&redemptionCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRedemptionCallbackArgs %v", err.Error()))
	}
	return out
}

func (k Keeper) UnmarshalRedemptionCallbackArgs(ctx sdk.Context, redemptionCallback []byte) types.RedemptionCallback {
	unmarshalledDelegateCallback := types.RedemptionCallback{}
	if err := proto.Unmarshal(redemptionCallback, &unmarshalledDelegateCallback); err != nil {
        k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRedemptionCallbackArgs %v", err.Error()))
	}
	return unmarshalledDelegateCallback
}

func RedemptionCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack []byte, args []byte) error {
	// QUESTION: should we check invariants here? e.g. sendMsg.FromAddress == redemptionAddress, msg type == MsgSend, etc.
	k.Logger(ctx).Info("RedemptionCallback executing", "packet", packet, "ack", ack, "args", args)
	// deserialize the args
	redemptionCallback := k.UnmarshalRedemptionCallbackArgs(ctx, args)
	k.Logger(ctx).Info(fmt.Sprintf("RedemptionCallback %v", redemptionCallback))
	userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, redemptionCallback.GetUserRedemptionRecordId())
	if !found {
		return sdkerrors.Wrap(types.ErrRecordNotFound, "user redemption record not found")
	}

	// deserialize the ack
	_, err := k.GetTxMsgData(ctx, ack)
	if err != nil {
		// ack failed, set UserRedemptionRecord as claimable
		// NOTE: we probably only want to do this if we could unmarshal the ack and it failed
		// DO NOT MERGE THIS INy

		userRedemptionRecord.IsClaimable = true
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
		return err
	}
	// claim successfully processed
	k.RecordsKeeper.RemoveUserRedemptionRecord(ctx, redemptionCallback.GetUserRedemptionRecordId())
	return nil
}


// ----------------------------------- helpers ----------------------------------- //
func (k Keeper) GetTxMsgData(ctx sdk.Context, acknowledgement []byte) (*sdk.TxMsgData, error) {
	ack := channeltypes.Acknowledgement_Result{}
	eventType := "callback"
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
		),
	)
	err := json.Unmarshal(acknowledgement, &ack)
	if err != nil {
		ackErr := channeltypes.Acknowledgement_Error{}
		err := json.Unmarshal(acknowledgement, &ackErr)
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					eventType,
					sdk.NewAttribute(types.AttributeKeyAckError, ackErr.Error),
				),
			)
			k.Logger(ctx).Error("Unable to unmarshal acknowledgement error", "error", err, "data", acknowledgement)
			return nil, err
		}
		k.Logger(ctx).Error("Unable to unmarshal acknowledgement result", "error", err, "remote_err", ackErr, "data", acknowledgement)
		return nil, err
	}

	txMsgData := &sdk.TxMsgData{}
	err = proto.Unmarshal(ack.Result, txMsgData)
	if err != nil {
		k.Logger(ctx).Error("Unable to unmarshal acknowledgement", "error", err, "ack", ack.Result)
		return nil, err
	}
	return txMsgData, nil
}
