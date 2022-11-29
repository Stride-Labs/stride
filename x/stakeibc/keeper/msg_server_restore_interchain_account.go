package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	recordtypes "github.com/Stride-Labs/stride/v3/x/records/types"
	"github.com/Stride-Labs/stride/v3/x/stakeibc/types"
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
		return nil, fmt.Errorf(`%s: %s`,errMsg, types.ErrInvalidInterchainAccountAddress.Error())
	}

	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, hostZone.ConnectionId, owner); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to register %s account : %s", msg.AccountType.String(), err))
		return nil, err
	}

	// If we're restoring a delegation account, we also have to reset record state
	if msg.AccountType == types.ICAAccountType_DELEGATION {
		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
		// revert DELEGATION_IN_PROGRESS records for the closed ICA channel (so that they can be staked)
		for _, record := range depositRecords {
			// only revert records for the select host zone
			if record.HostZoneId == hostZone.ChainId && record.Status == recordtypes.DepositRecord_DELEGATION_IN_PROGRESS {
				record.Status = recordtypes.DepositRecord_DELEGATION_QUEUE
				k.Logger(ctx).Error(fmt.Sprintf("Setting DepositRecord %d to status DepositRecord_DELEGATION_IN_PROGRESS", record.Id))
				k.RecordsKeeper.SetDepositRecord(ctx, record)
			}
		}
	}

	return &types.MsgRestoreInterchainAccountResponse{}, nil
}
