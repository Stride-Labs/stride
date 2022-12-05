package keeper

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
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

func UndelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	logMsg := fmt.Sprintf("UndelegateCallback executing packet: %d, source: %s %s, dest: %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
	k.Logger(ctx).Info(logMsg)

	// fetch relevant state
	undelegateCallback, err := k.UnmarshalUndelegateCallbackArgs(ctx, args)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal undelegate callback args | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("UndelegateCallback, HostZone: %s", undelegateCallback.HostZoneId))
	zone, found := k.GetHostZone(ctx, undelegateCallback.HostZoneId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", undelegateCallback.HostZoneId)
	}

	// handle transaction failure cases
	if ack == nil {
		// handle timeout
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback timeout, txMsgData is nil, packet %v", packet))
		return nil
	}
	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}
	if len(txMsgData.Data) == 0 {
		// handle tx failure
		// reset to UNBONDING_QUEUE
		err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, zone.ChainId, undelegateCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_UNBONDING_QUEUE)
		if err != nil {
			return err
		}
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback tx failed, txMsgData is empty, ack error, packet %v", packet))
		return nil
	}

	// core callback logic
	err = k.UpdateDelegationBalances(ctx, zone, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}
	latestCompletionTime, err := k.GetLatestCompletionTime(ctx, txMsgData)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}
	stTokenBurnAmount, err := k.UpdateHostZoneUnbondings(ctx, *latestCompletionTime, zone, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}
	err = k.BurnTokens(ctx, zone, stTokenBurnAmount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}
	// upon success, add host zone unbondings to the exit transfer queue
	err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, zone.ChainId, undelegateCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateDelegationBalances(ctx sdk.Context, zone types.HostZone, undelegateCallback types.UndelegateCallback) error {
	// Undelegate from each validator and update host zone staked balance, if successful
	for _, undelegation := range undelegateCallback.SplitDelegations {
		undelegateAmt, err := cast.ToInt64E(undelegation.Amount)
		k.Logger(ctx).Info(fmt.Sprintf("UndelegateCallback, Undelegation: %d, validator: %s", undelegateAmt, undelegation.Validator))
		if err != nil {
			errMsg := fmt.Sprintf("Could not convert undelegate amount to int64 in undelegation callback | %s", err.Error())
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrIntCast, errMsg)
		}
		undelegateVal := undelegation.Validator
		success := k.AddDelegationToValidator(ctx, zone, undelegateVal, -undelegateAmt)
		if !success {
			return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to remove delegation to validator")
		}
		if undelegation.Amount > zone.StakedBal {
			// handle incoming underflow
			// Once we add a killswitch, we should also stop liquid staking on the zone here
			return sdkerrors.Wrapf(types.ErrUndelegationAmount, "undelegation.Amount > zone.StakedBal, undelegation.Amount: %d, zone.StakedBal %d", undelegation.Amount, zone.StakedBal)
		} else {
			zone.StakedBal -= undelegation.Amount
		}
	}
	k.SetHostZone(ctx, zone)
	return nil
}

func (k Keeper) GetLatestCompletionTime(ctx sdk.Context, txMsgData *sdk.TxMsgData) (*time.Time, error) {
	// Update the completion time using the latest completion time across each message within the transaction
	latestCompletionTime := time.Time{}
	for _, msgResponseBytes := range txMsgData.Data {
		var undelegateResponse stakingtypes.MsgUndelegateResponse
		if msgResponseBytes == nil || msgResponseBytes.Data == nil {
			return nil, sdkerrors.Wrap(types.ErrTxMsgDataInvalid, "msgResponseBytes or msgResponseBytes.Data is nil")
		}
		err := proto.Unmarshal(msgResponseBytes.Data, &undelegateResponse)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to unmarshal undelegation tx response | %s", err)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
		}
		if undelegateResponse.CompletionTime.After(latestCompletionTime) {
			latestCompletionTime = undelegateResponse.CompletionTime
		}
	}
	if latestCompletionTime.IsZero() {
		errMsg := fmt.Sprintf("Invalid completion time (%s) from txMsg", latestCompletionTime.String())
		k.Logger(ctx).Error(errMsg)
		return nil, types.ErrInvalidPacketCompletionTime
	}
	return &latestCompletionTime, nil
}

func (k Keeper) UpdateHostZoneUnbondings(
	ctx sdk.Context,
	latestCompletionTime time.Time,
	zone types.HostZone,
	undelegateCallback types.UndelegateCallback,
) (stTokenBurnAmount int64, err error) {
	// UpdateHostZoneUnbondings does two things:
	// 		1. Update the status and time of each hostZoneUnbonding on each epochUnbondingRecord
	// 		2. Return the number of stTokens that need to be burned
	for _, epochNumber := range undelegateCallback.EpochUnbondingRecordIds {
		epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
		if !found {
			errMsg := fmt.Sprintf("Unable to find epoch unbonding record for epoch: %d", epochNumber)
			k.Logger(ctx).Error(errMsg)
			return 0, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, zone.ChainId)
		if !found {
			errMsg := fmt.Sprintf("Host zone unbonding not found (%s) in epoch unbonding record: %d", zone.ChainId, epochNumber)
			k.Logger(ctx).Error(errMsg)
			return 0, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}

		// Keep track of the stTokens that need to be burned
		stTokenAmount, err := cast.ToInt64E(hostZoneUnbonding.StTokenAmount)
		if err != nil {
			errMsg := fmt.Sprintf("Could not convert stTokenAmount to int64 in redeem stake | %s", err.Error())
			k.Logger(ctx).Error(errMsg)
			return 0, sdkerrors.Wrapf(types.ErrIntCast, errMsg)
		}
		stTokenBurnAmount += stTokenAmount

		// Update the bonded status and time
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
		hostZoneUnbonding.UnbondingTime = cast.ToUint64(latestCompletionTime.UnixNano())
		updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, zone.ChainId, hostZoneUnbonding)
		if !success {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v", epochUnbondingRecord.EpochNumber, zone.ChainId, hostZoneUnbonding))
			return 0, sdkerrors.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
		}
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

		logMsg := fmt.Sprintf("Set unbonding time to %s for host zone %s's unbonding record: %d",
			latestCompletionTime.String(), zone.ChainId, epochNumber)
		k.Logger(ctx).Info(logMsg)
	}
	return stTokenBurnAmount, nil
}

func (k Keeper) BurnTokens(ctx sdk.Context, zone types.HostZone, stTokenBurnAmount int64) error {
	stCoinDenom := types.StAssetDenomFromHostZoneDenom(zone.HostDenom)
	stCoinString := sdk.NewDec(stTokenBurnAmount).String() + stCoinDenom
	stCoin, err := sdk.ParseCoinNormalized(stCoinString)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse burnCoin: %s. err: %s", stCoinString, err.Error())
	}
	bech32ZoneAddress, err := sdk.AccAddressFromBech32(zone.Address)
	if err != nil {
		return fmt.Errorf("could not bech32 decode address %s of zone with id: %s", zone.Address, zone.ChainId)
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, bech32ZoneAddress, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		return fmt.Errorf("could not send coins from account %s to module %s. err: %s", zone.Address, types.ModuleName, err.Error())
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to burn stAssets upon successful unbonding %s", err.Error()))
		return sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't burn %d %s tokens in module account. err: %s", stTokenBurnAmount, stCoinDenom, err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, stCoinDenom)))
	return nil
}
