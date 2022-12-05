package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k msgServer) RestoreInterchainAccount(goCtx context.Context, msg *types.MsgRestoreInterchainAccount) (*types.MsgRestoreInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	owner := types.FormatICAAccountOwner(msg.ChainId, msg.AccountType)

	// only allow restoring an account if it already exists
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		errMsg := fmt.Sprintf("could not create portID for ICA controller account address: %s", owner)
		k.Logger(ctx).Error(errMsg)
		return nil, err
	}
	_, exists := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, hostZone.ConnectionId, portID)
	if !exists {
		errMsg := fmt.Sprintf("ICA controller account address not found: %s", owner)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrInvalidInterchainAccountAddress, errMsg)
	}

	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, hostZone.ConnectionId, owner); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to register %s account : %s", msg.AccountType.String(), err))
		return nil, err
	}

	// If we're restoring a delegation account, we also have to reset record state
	if msg.AccountType == types.ICAAccountType_DELEGATION {
		// revert DELEGATION_IN_PROGRESS records for the closed ICA channel (so that they can be staked)
		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
		for _, depositRecord := range depositRecords {
			// only revert records for the select host zone
			if depositRecord.HostZoneId == hostZone.ChainId && depositRecord.Status == recordtypes.DepositRecord_DELEGATION_IN_PROGRESS {
				depositRecord.Status = recordtypes.DepositRecord_DELEGATION_QUEUE
				k.Logger(ctx).Info(fmt.Sprintf("Setting DepositRecord %d to status DepositRecord_DELEGATION_IN_PROGRESS", depositRecord.Id))
				k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
			}
		}

		// revert epoch unbonding records for the closed ICA channel
		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
		epochNumberForPendingUnbondingRecords := []uint64{}
		epochNumberForPendingTransferRecords := []uint64{}
		for _, epochUnbondingRecord := range epochUnbondingRecords {
			// only revert records for the select host zone
			hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
			if !found {
				k.Logger(ctx).Info(fmt.Sprintf("No HostZoneUnbonding found for chainId: %s, epoch: %d", hostZone.ChainId, epochUnbondingRecord.EpochNumber))
				continue
			}

			// Revert UNBONDING_IN_PROGRESS and EXIT_TRANSFER_IN_PROGRESS records
			if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d is stuck in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS.String(),
				))
				epochNumberForPendingUnbondingRecords = append(epochNumberForPendingUnbondingRecords, epochUnbondingRecord.EpochNumber)

			} else if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d to in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS.String(),
				))
				epochNumberForPendingTransferRecords = append(epochNumberForPendingTransferRecords, epochUnbondingRecord.EpochNumber)
			}
		}
		// Revert UNBONDING_IN_PROGRESS records to UNBONDING_QUEUE
		err := k.RecordsKeeper.SetHostZoneUnbondings(ctx, hostZone.ChainId, epochNumberForPendingUnbondingRecords, recordtypes.HostZoneUnbonding_UNBONDING_QUEUE)
		if err != nil {
			errMsg := fmt.Sprintf("unable to update host zone unbonding record status to %s for chainId: %s and epochUnbondingRecordIds: %v, err: %s",
				recordtypes.HostZoneUnbonding_UNBONDING_QUEUE.String(), hostZone.ChainId, epochNumberForPendingUnbondingRecords, err)
			k.Logger(ctx).Error(errMsg)
			return nil, err
		}

		// Revert EXIT_TRANSFER_IN_PROGRESS records to EXIT_TRANSFER_QUEUE
		err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, hostZone.ChainId, epochNumberForPendingTransferRecords, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
		if err != nil {
			errMsg := fmt.Sprintf("unable to update host zone unbonding record status to %s for chainId: %s and epochUnbondingRecordIds: %v, err: %s",
				recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE.String(), hostZone.ChainId, epochNumberForPendingTransferRecords, err)
			k.Logger(ctx).Error(errMsg)
			return nil, err
		}
	}

	return &types.MsgRestoreInterchainAccountResponse{}, nil
}
