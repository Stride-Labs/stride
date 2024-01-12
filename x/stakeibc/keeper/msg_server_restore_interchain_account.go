package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"

	recordtypes "github.com/Stride-Labs/stride/v17/x/records/types"
	"github.com/Stride-Labs/stride/v17/x/stakeibc/types"
)

func (k msgServer) RestoreInterchainAccount(goCtx context.Context, msg *types.MsgRestoreInterchainAccount) (*types.MsgRestoreInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get ConnectionEnd (for counterparty connection)
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)
	if !found {
		return nil, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection %s not found", msg.ConnectionId)
	}
	counterpartyConnection := connectionEnd.Counterparty

	// only allow restoring an account if it already exists
	portID, err := icatypes.NewControllerPortID(msg.AccountOwner)
	if err != nil {
		return nil, err
	}
	_, exists := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, msg.ConnectionId, portID)
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrInvalidInterchainAccountAddress,
			"ICA controller account address not found: %s", msg.AccountOwner)
	}

	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: msg.ConnectionId,
		HostConnectionId:       counterpartyConnection.ConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.AccountOwner, appVersion); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register account for owner %s", msg.AccountOwner)
	}

	// If we're restoring a delegation account, we also have to reset record state
	if msg.AccountOwner == types.FormatHostZoneICAOwner(msg.ChainId, types.ICAAccountType_DELEGATION) {
		hostZone, found := k.GetHostZone(ctx, msg.ChainId)
		if !found {
			return nil, types.ErrHostZoneNotFound.Wrapf("delegation ICA supplied, but no associated host zone")
		}

		// Since any ICAs along the original channel will never get relayed,
		// we have to reset the delegation_changes_in_progress field on each validator
		for _, validator := range hostZone.Validators {
			validator.DelegationChangesInProgress = 0
		}
		k.SetHostZone(ctx, hostZone)

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
		err := k.RecordsKeeper.SetHostZoneUnbondingStatus(ctx, hostZone.ChainId, epochNumberForPendingUnbondingRecords, recordtypes.HostZoneUnbonding_UNBONDING_QUEUE)
		if err != nil {
			errMsg := fmt.Sprintf("unable to update host zone unbonding record status to %s for chainId: %s and epochUnbondingRecordIds: %v, err: %s",
				recordtypes.HostZoneUnbonding_UNBONDING_QUEUE.String(), hostZone.ChainId, epochNumberForPendingUnbondingRecords, err)
			k.Logger(ctx).Error(errMsg)
			return nil, err
		}

		// Revert EXIT_TRANSFER_IN_PROGRESS records to EXIT_TRANSFER_QUEUE
		err = k.RecordsKeeper.SetHostZoneUnbondingStatus(ctx, hostZone.ChainId, epochNumberForPendingTransferRecords, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
		if err != nil {
			errMsg := fmt.Sprintf("unable to update host zone unbonding record status to %s for chainId: %s and epochUnbondingRecordIds: %v, err: %s",
				recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE.String(), hostZone.ChainId, epochNumberForPendingTransferRecords, err)
			k.Logger(ctx).Error(errMsg)
			return nil, err
		}

		// Revert all pending LSM Detokenizations from status DETOKENIZATION_IN_PROGRESS to status DETOKENIZATION_QUEUE
		pendingDeposits := k.RecordsKeeper.GetLSMDepositsForHostZoneWithStatus(ctx, hostZone.ChainId, recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS)
		for _, lsmDeposit := range pendingDeposits {
			k.Logger(ctx).Info(fmt.Sprintf("Setting LSMTokenDeposit %s to status DETOKENIZATION_QUEUE", lsmDeposit.Denom))
			k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, lsmDeposit, recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE)
		}
	}

	return &types.MsgRestoreInterchainAccountResponse{}, nil
}
