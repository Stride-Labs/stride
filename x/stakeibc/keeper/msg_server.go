package keeper

import (
	"context"
	"fmt"

	proto "github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v11/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v34/utils"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v34/x/records/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (ms msgServer) UpdateHostZoneParams(goCtx context.Context, msg *types.MsgUpdateHostZoneParams) (*types.MsgUpdateHostZoneParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	hostZone, found := ms.Keeper.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	maxMessagesPerTx := msg.MaxMessagesPerIcaTx
	if maxMessagesPerTx == 0 {
		maxMessagesPerTx = DefaultMaxMessagesPerIcaTx
	}
	hostZone.MaxMessagesPerIcaTx = maxMessagesPerTx
	ms.Keeper.SetHostZone(ctx, hostZone)

	return &types.MsgUpdateHostZoneParamsResponse{}, nil
}

// Gov transaction to deprecate a host zone
func (ms msgServer) DeprecateHostZone(goCtx context.Context, msg *types.MsgDeprecateHostZone) (*types.MsgDeprecateHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	hostZone, found := ms.Keeper.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	// The halted is set to freeze business logic, deprecated is just used as an annotation/documentation
	// but does not impact anything functionally
	hostZone.Halted = true
	hostZone.Deprecated = true
	ms.Keeper.SetHostZone(ctx, hostZone)

	return &types.MsgDeprecateHostZoneResponse{}, nil
}

func (k msgServer) AddValidators(goCtx context.Context, msg *types.MsgAddValidators) (*types.MsgAddValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, false); err != nil {
			return nil, err
		}

		// Query and store the validator's sharesToTokens rate
		if err := k.QueryValidatorSharesToTokensRate(ctx, msg.HostZone, validator.Address); err != nil {
			return nil, err
		}
	}

	// Confirm none of the validator's exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, msg.HostZone); err != nil {
		return nil, err
	}

	return &types.MsgAddValidatorsResponse{}, nil
}

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to remove validator %s from host zone %s", msg.ValAddr, msg.HostZone)
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}

func (k msgServer) ChangeValidatorWeight(goCtx context.Context, msg *types.MsgChangeValidatorWeights) (*types.MsgChangeValidatorWeightsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, types.ErrInvalidHostZone
	}

	for _, weightChange := range msg.ValidatorWeights {

		validatorFound := false
		for _, validator := range hostZone.Validators {
			if validator.Address == weightChange.Address {
				validator.Weight = weightChange.Weight
				k.SetHostZone(ctx, hostZone)

				validatorFound = true
				break
			}
		}

		if !validatorFound {
			return nil, types.ErrValidatorNotFound
		}
	}

	// Confirm the new weights wouldn't cause any validator to exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, msg.HostZone); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to change validator weight")
	}

	return &types.MsgChangeValidatorWeightsResponse{}, nil
}

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

	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.AccountOwner, appVersion, channeltypes.ORDERED); err != nil {
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
				depositRecord.DelegationTxsInProgress = 0

				k.Logger(ctx).Info(fmt.Sprintf("Setting DepositRecord %d to status DepositRecord_DELEGATION_IN_PROGRESS", depositRecord.Id))
				k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
			}
		}

		// a pending undelegation batch in flight on the dead channel can never be acked (and its
		// timeout may never be processable), so release it for the next day epoch to resubmit
		k.RemovePendingUndelegationInFlight(ctx, hostZone.ChainId)

		// revert epoch unbonding records for the closed ICA channel
		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
		for _, epochUnbondingRecord := range epochUnbondingRecords {
			// only revert records for the select host zone
			hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
			if !found {
				k.Logger(ctx).Info(fmt.Sprintf("No HostZoneUnbonding found for chainId: %s, epoch: %d", hostZone.ChainId, epochUnbondingRecord.EpochNumber))
				continue
			}

			// Reset the number of undelegation txs in progress
			hostZoneUnbonding.UndelegationTxsInProgress = 0

			// Revert UNBONDING_IN_PROGRESS records to UNBONDING_RETRY_QUEUE
			// and EXIT_TRANSFER_IN_PROGRESS records to EXIT_TRANSFER_QUEUE
			if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d is stuck in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS.String(),
				))
				hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE

			} else if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d to in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS.String(),
				))
				hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
			}

			err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId, *hostZoneUnbonding)
			if err != nil {
				return nil, err
			}
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

// Admin transaction to close an ICA channel by sending an ICA with a 1 nanosecond timeout (which will force a timeout and closure)
// This can be used if there are records stuck in state IN_PROGRESS after a channel has been re-opened after a timeout
// After the closure, the a new channel can be permissionlessly re-opened with RestoreInterchainAccount
func (k msgServer) CloseDelegationChannel(goCtx context.Context, msg *types.MsgCloseDelegationChannel) (*types.MsgCloseDelegationChannelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound.Wrapf("chain id %s", msg.ChainId)
	}

	// Submit an ICA bank send from the delegation ICA account to itself for just 1utoken
	delegationIcaOwner := types.FormatHostZoneICAOwner(msg.ChainId, types.ICAAccountType_DELEGATION)
	msgSend := []proto.Message{&banktypes.MsgSend{
		FromAddress: hostZone.DelegationIcaAddress,
		ToAddress:   hostZone.DelegationIcaAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin(hostZone.HostDenom, sdkmath.OneInt())),
	}}

	// Timeout the ICA 1 nanosecond after the current block time (so it's impossible to be relayed)
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().UnixNano() + 1)
	err := k.SubmitICATxWithoutCallback(ctx, hostZone.ConnectionId, delegationIcaOwner, msgSend, timeoutTimestamp)
	if err != nil {
		return nil, err
	}

	return &types.MsgCloseDelegationChannelResponse{}, nil
}

// This kicks off two ICQs, each with a callback, that will update the number of tokens on a validator
// after being slashed. The flow is:
// 1. QueryValidatorSharesToTokensRate (ICQ)
// 2. ValidatorSharesToTokensRate (CALLBACK)
// 3. SubmitDelegationICQ (ICQ)
// 4. DelegatorSharesCallback (CALLBACK)
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.QueryValidatorSharesToTokensRate(ctx, msg.ChainId, msg.Valoper); err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}

// Submits an ICQ to get the validator's delegated shares
func (k msgServer) CalibrateDelegation(goCtx context.Context, msg *types.MsgCalibrateDelegation) (*types.MsgCalibrateDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound
	}

	if err := k.SubmitCalibrationICQ(ctx, hostZone, msg.Valoper); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return nil, err
	}

	return &types.MsgCalibrateDelegationResponse{}, nil
}

func (k msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: we're intentionally not checking the zone is halted
	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	// Get the wide bounds
	outerMinSafetyThreshold, outerMaxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	innerMinSafetyThreshold := msg.MinInnerRedemptionRate
	innerMaxSafetyThreshold := msg.MaxInnerRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	if innerMinSafetyThreshold.LT(outerMinSafetyThreshold) {
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds,
			"inner min safety threshold (%s) is less than outer min safety threshold (%s)",
			innerMinSafetyThreshold, outerMinSafetyThreshold)
	}

	if innerMaxSafetyThreshold.GT(outerMaxSafetyThreshold) {
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds,
			"inner max safety threshold (%s) is greater than outer max safety threshold (%s)",
			innerMaxSafetyThreshold, outerMaxSafetyThreshold)
	}

	// Set the inner bounds on the host zone
	zone.MinInnerRedemptionRate = innerMinSafetyThreshold
	zone.MaxInnerRedemptionRate = innerMaxSafetyThreshold

	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}

func (k msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get Host Zone
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone %s not found", msg.ChainId)
	}

	// Check the zone is halted
	if !hostZone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotHalted, "host zone %s is not halted", msg.ChainId)
	}

	// remove from blacklist
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	k.RatelimitKeeper.RemoveDenomFromBlacklist(ctx, stDenom)

	// Resume zone
	hostZone.Halted = false
	k.SetHostZone(ctx, hostZone)

	return &types.MsgResumeHostZoneResponse{}, nil
}
