package keeper

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"
)

func (k Keeper) MarshalUndelegateCallbackArgs(ctx sdk.Context, undelegateCallback types.UndelegateCallback) ([]byte, error) {
	out, err := proto.Marshal(&undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalUndelegateCallbackArgs | %s", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalUndelegateCallbackArgs(ctx sdk.Context, undelegateCallback []byte) (types.UndelegateCallback, error) {
	unmarshalledUndelegateCallback := types.UndelegateCallback{}
	if err := proto.Unmarshal(undelegateCallback, &unmarshalledUndelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalUndelegateCallbackArgs | %s", err.Error()))
		return unmarshalledUndelegateCallback, err
	}
	return unmarshalledUndelegateCallback, nil
}

func UndelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement_Result, args []byte) error {
	logMsg := fmt.Sprintf("UndelegateCallback executing packet: %d, source: %s %s, dest: %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
	k.Logger(ctx).Info(logMsg)

	if ack == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "ack is nil")
	}

	// unmarshal the callback args and get the host zone
	undelegateCallback, err := k.UnmarshalUndelegateCallbackArgs(ctx, args)
	k.Logger(ctx).Info(fmt.Sprintf("UndelegateCallback, HostZone: %s", undelegateCallback.HostZoneId))

	hostZone, found := k.GetHostZone(ctx, undelegateCallback.HostZoneId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", undelegateCallback.HostZoneId)
	}

	// Redelegate to each validator and updated host zone staked balance if successful
	for _, undelegation := range undelegateCallback.SplitDelegations {
		undelegateAmt := int64(undelegation.Amount)
		undelegateVal := undelegation.Validator
		success := k.AddDelegationToValidator(ctx, hostZone, undelegateVal, -undelegateAmt)
		if !success {
			return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to add delegation to validator")
		}
		hostZone.StakedBal -= undelegateAmt
	}
	k.SetHostZone(ctx, hostZone)

	// Get the individual msg responses from inside the transaction
	txMsgData := &sdk.TxMsgData{}
	err = proto.Unmarshal(ack.Result, txMsgData)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalUndelegateCallbackArgs | %s", err.Error()))
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, "undelegate callback args")
	}

	// Update the completion time using the latest completion time across each message within the transaction
	latestCompletionTime := time.Time{}
	for _, msgResponseBytes := range txMsgData.Data {
		var undelegateResponse stakingtypes.MsgUndelegateResponse
		err := proto.Unmarshal(msgResponseBytes.Data, &undelegateResponse)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to unmarshal undelegation tx response | %s", err)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
		}
		if undelegateResponse.CompletionTime.After(latestCompletionTime) {
			latestCompletionTime = undelegateResponse.CompletionTime
		}
	}
	if latestCompletionTime.IsZero() {
		errMsg := fmt.Sprintf("Invalid completion time from txMsg (%s)", latestCompletionTime.String())
		k.Logger(ctx).Error(errMsg)
		return types.ErrInvalidPacketCompletionTime
	}

	// Update the status and time of each unbonding record and grab the number of stTokens that need to be burned
	stAmountToBurn := int64(0)
	for _, epochNumber := range undelegateCallback.UnbondingEpochNumbers {
		epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
		if !found {
			errMsg := fmt.Sprintf("could find epoch unbonding record for epoch: %d", epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}
		hostZoneUnbondings, found := epochUnbondingRecord.HostZoneUnbondings[hostZone.ChainId]
		if !found {
			errMsg := fmt.Sprintf("host zone not found (%s) in epoch unbonding record: %d", hostZone.ChainId, epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}

		// Keep track of the stTokens that need to be burned
		stAmountToBurn += int64(hostZoneUnbondings.StTokenAmount)

		// Update the bonded status and time
		hostZoneUnbondings.Status = recordstypes.HostZoneUnbonding_UNBONDED
		hostZoneUnbondings.UnbondingTime = cast.ToUint64(latestCompletionTime.UnixNano())
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)

		logMsg := fmt.Sprintf("Set unbonding time to %s for host zone %s's unbonding record: %d",
			latestCompletionTime.String(), hostZone.ChainId, epochNumber)
		k.Logger(ctx).Info(logMsg)
	}

	// Burn stTokens
	stCoinDenom := "st" + hostZone.HostDenom
	stCoinString := sdk.NewDec(stAmountToBurn).String() + stCoinDenom
	stCoin, err := sdk.ParseCoinNormalized(stCoinString)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse burnCoin: %s. err: %s", stCoinString, err.Error())
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to burn stAssets upon successful unbonding %v", err))
		return sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't burn %d %s tokens in module account. err: %s", stAmountToBurn, stCoinDenom, err.Error())
	}

	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, "stuatom")))
	return nil
}
