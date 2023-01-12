package keeper

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
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

// Marshal undelegate callback args
func (k Keeper) MarshalUndelegateCallbackArgs(ctx sdk.Context, undelegateCallback types.UndelegateCallback) ([]byte, error) {
	out, err := proto.Marshal(&undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalUndelegateCallbackArgs | %s", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls undelegate callback arguments into a UndelegateCallback struct
func (k Keeper) UnmarshalUndelegateCallbackArgs(ctx sdk.Context, undelegateCallback []byte) (types.UndelegateCallback, error) {
	unmarshalledUndelegateCallback := types.UndelegateCallback{}
	if err := proto.Unmarshal(undelegateCallback, &unmarshalledUndelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalUndelegateCallbackArgs | %s", err.Error()))
		return unmarshalledUndelegateCallback, err
	}
	return unmarshalledUndelegateCallback, nil
}

// ICA Callback after undelegating
//   If successful:
//      * Updates epoch unbonding record status
// 		* Records delegation changes on the host zone and validators,
//      * Burns stTokens
//   If timeout:
//      * Does nothing
//   If failure:
//		* Reverts epoch unbonding record status
func UndelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	// Fetch callback args
	undelegateCallback, err := k.UnmarshalUndelegateCallbackArgs(ctx, args)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal undelegate callback args | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
	}
	chainId := undelegateCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Undelegate,
		"Starting callback for Epoch Unbonding Records: %+v", undelegateCallback.EpochUnbondingRecordIds))

	// Check for timeout (ack nil)
	// No need to reset the unbonding record status since it will get revertted when the channel is restored
	if ack == nil {
		k.Logger(ctx).Error(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Undelegate,
			"TIMEOUT (ack is nil), Packet: %+v", packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// Reset the unbonding record status upon failure
	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}
	if len(txMsgData.Data) == 0 {
		k.Logger(ctx).Error(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Undelegate,
			"ICA TX FAILED (ack is empty / ack error), Packet: %+v", packet))

		// Reset unbondings record status
		err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, chainId, undelegateCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_UNBONDING_QUEUE)
		if err != nil {
			return err
		}
		return nil
	}

	k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Undelegate, "SUCCESS, Packet: %+v", packet))

	// Update delegation balances
	hostZone, found := k.GetHostZone(ctx, undelegateCallback.HostZoneId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", undelegateCallback.HostZoneId)
	}
	err = k.UpdateDelegationBalances(ctx, hostZone, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Get the latest transaction completion time (to determine the unbonding time)
	latestCompletionTime, err := k.GetLatestCompletionTime(ctx, txMsgData)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Burn the stTokens
	stTokenBurnAmount, err := k.UpdateHostZoneUnbondings(ctx, *latestCompletionTime, chainId, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}
	err = k.BurnTokens(ctx, hostZone, stTokenBurnAmount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Upon success, add host zone unbondings to the exit transfer queue
	err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, chainId, undelegateCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
	if err != nil {
		return err
	}

	return nil
}

// Decrement the stakedBal field on the host zone and each validator's delegations after a successful unbonding ICA
func (k Keeper) UpdateDelegationBalances(ctx sdk.Context, zone types.HostZone, undelegateCallback types.UndelegateCallback) error {
	// Undelegate from each validator and update host zone staked balance, if successful
	for _, undelegation := range undelegateCallback.SplitDelegations {
		success := k.AddDelegationToValidator(ctx, zone, undelegation.Validator, undelegation.Amount.Neg(), ICACallbackID_Undelegate)
		if !success {
			return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to remove delegation to validator")
		}

		if undelegation.Amount.GT(zone.StakedBal) {
			// handle incoming underflow
			// Once we add a killswitch, we should also stop liquid staking on the zone here
			return sdkerrors.Wrapf(types.ErrUndelegationAmount, "undelegation.Amount > zone.StakedBal, undelegation.Amount: %v, zone.StakedBal %v", undelegation.Amount, zone.StakedBal)
		} else {
			zone.StakedBal = zone.StakedBal.Sub(undelegation.Amount)
		}
	}
	k.SetHostZone(ctx, zone)
	return nil
}

// Get the latest completion time across each MsgUndelegate in the ICA transaction
// The time is used to set the
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

// UpdateHostZoneUnbondings does two things:
// 		1. Update the status and time of each hostZoneUnbonding on each epochUnbondingRecord
// 		2. Return the number of stTokens that need to be burned
func (k Keeper) UpdateHostZoneUnbondings(
	ctx sdk.Context,
	latestCompletionTime time.Time,
	chainId string,
	undelegateCallback types.UndelegateCallback,
) (stTokenBurnAmount sdk.Int, err error) {
	stTokenBurnAmount = sdk.ZeroInt()
	for _, epochNumber := range undelegateCallback.EpochUnbondingRecordIds {
		epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
		if !found {
			errMsg := fmt.Sprintf("Unable to find epoch unbonding record for epoch: %d", epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdk.ZeroInt(), sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, chainId)
		if !found {
			errMsg := fmt.Sprintf("Host zone unbonding not found (%s) in epoch unbonding record: %d", chainId, epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdk.ZeroInt(), sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}

		// Keep track of the stTokens that need to be burned
		stTokenAmount := hostZoneUnbonding.StTokenAmount
		stTokenBurnAmount = stTokenBurnAmount.Add(stTokenAmount)

		// Update the bonded status and time
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
		hostZoneUnbonding.UnbondingTime = cast.ToUint64(latestCompletionTime.UnixNano())
		updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding)
		if !success {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %+v",
				epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding))
			return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
		}
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

		k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Undelegate,
			"Epoch Unbonding Record: %d - Seting unbonding time to %s", epochNumber, latestCompletionTime.String()))
	}
	return stTokenBurnAmount, nil
}

// Burn stTokens after they've been unbonded
func (k Keeper) BurnTokens(ctx sdk.Context, hostZone types.HostZone, stTokenBurnAmount sdk.Int) error {
	// Build the coin from the stDenom on the host zone
	stCoinDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stCoinString := stTokenBurnAmount.String() + stCoinDenom
	stCoin, err := sdk.ParseCoinNormalized(stCoinString)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse burnCoin: %s. err: %s", stCoinString, err.Error())
	}

	// Send the stTokens from the host zone module account to the stakeibc module account
	bech32ZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.Address, hostZone.ChainId)
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, bech32ZoneAddress, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		return fmt.Errorf("could not send coins from account %s to module %s. err: %s", hostZone.Address, types.ModuleName, err.Error())
	}

	// Finally burn the stTokens
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to burn stAssets upon successful unbonding %s", err.Error()))
		return sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't burn %v%s tokens in module account. err: %s", stTokenBurnAmount, stCoinDenom, err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, stCoinDenom)))
	return nil
}
