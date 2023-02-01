package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	recordstypes "github.com/Stride-Labs/stride/v5/x/records/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
)

func (k Keeper) DelegateOnHost(ctx sdk.Context, hostZone types.HostZone, amt sdk.Coin, depositRecord recordstypes.DepositRecord) error {
	// the relevant ICA is the delegate account
	owner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return fmt.Errorf("%s has no associated portId: %s", owner, types.ErrInvalidAddress.Error())
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return fmt.Errorf("%s has no associated connection: %s", portID, types.ErrInvalidChainID.Error())
	}

	// Fetch the relevant ICA
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return fmt.Errorf("Invalid delegation account: %s", types.ErrInvalidAddress.Error())
	}

	// Construct the transaction
	targetDelegatedAmts, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, amt.Amount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target delegation amounts for host zone %s", hostZone.ChainId))
		return err
	}

	var splitDelegations []*types.SplitDelegation
	var msgs []sdk.Msg
	for _, validator := range hostZone.Validators {
		relativeAmount := sdk.NewCoin(amt.Denom, targetDelegatedAmts[validator.Address])
		if relativeAmount.Amount.IsPositive() {
			msgs = append(msgs, &stakingTypes.MsgDelegate{
				DelegatorAddress: delegationAccount.Address,
				ValidatorAddress: validator.Address,
				Amount:           relativeAmount,
			})
		}
		splitDelegations = append(splitDelegations, &types.SplitDelegation{
			Validator: validator.Address,
			Amount:    relativeAmount.Amount,
		})
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Preparing MsgDelegates from the delegation account to each validator"))

	// add callback data
	delegateCallback := types.DelegateCallback{
		HostZoneId:       hostZone.ChainId,
		DepositRecordId:  depositRecord.Id,
		SplitDelegations: splitDelegations,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Marshalling DelegateCallback args: %+v", delegateCallback))
	marshalledCallbackArgs, err := k.MarshalDelegateCallbackArgs(ctx, delegateCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationAccount, ICACallbackID_Delegate, marshalledCallbackArgs)
	if err != nil {
		return fmt.Errorf("Failed to SubmitTxs for connectionId %s on %s. Messages: %s: %s", connectionId, hostZone.ChainId, msgs, err.Error())
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgDelegates Successfully Sent"))

	// update the record state to DELEGATION_IN_PROGRESS
	depositRecord.Status = recordstypes.DepositRecord_DELEGATION_IN_PROGRESS
	k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)

	return nil
}

func (k Keeper) SetWithdrawalAddressOnHost(ctx sdk.Context, hostZone types.HostZone) error {
	// The relevant ICA is the delegate account
	owner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return fmt.Errorf("%s has no associated portId: %s", owner, types.ErrInvalidAddress.Error())
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return fmt.Errorf("%s has no associated connection: %s", portID, types.ErrInvalidChainID.Error())
	}

	// Fetch the relevant ICA
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return nil
	}
	withdrawalAccount := hostZone.WithdrawalAccount
	if withdrawalAccount == nil || withdrawalAccount.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a withdrawal address!", hostZone.ChainId))
		return nil
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Withdrawal Address: %s, Delegator Address: %s",
		withdrawalAccount.Address, delegationAccount.Address))

	// Construct the ICA message
	msgs := []sdk.Msg{
		&distributiontypes.MsgSetWithdrawAddress{
			DelegatorAddress: delegationAccount.Address,
			WithdrawAddress:  withdrawalAccount.Address,
		},
	}
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationAccount, "", nil)
	if err != nil {
		return fmt.Errorf("Failed to SubmitTxs for %s, %s, %s: %s", connectionId, hostZone.ChainId, msgs, types.ErrInvalidRequest.Error())
	}

	return nil
}

// Submits an ICQ for the withdrawal account balance
func (k Keeper) UpdateWithdrawalBalance(ctx sdk.Context, hostZone types.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for withdrawal account balance"))

	// Get the withdrawal account address from the host zone
	withdrawalAccount := hostZone.WithdrawalAccount
	if withdrawalAccount == nil || withdrawalAccount.Address == "" {
		return fmt.Errorf("no withdrawal account found for %s: %s", hostZone.ChainId, types.ErrICAAccountNotFound.Error())
	}

	// Encode the withdrawal account address for the query request
	// The query request consists of the withdrawal account address and denom
	_, withdrawalAddressBz, err := bech32.DecodeAndConvert(withdrawalAccount.Address)
	if err != nil {
		return fmt.Errorf("invalid withdrawal account address, could not decode (%s): %s", err.Error(), types.ErrInvalidRequest.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(withdrawalAddressBz), []byte(hostZone.HostDenom)...)

	// The query should timeout at the end of the ICA buffer window
	ttl, err := k.GetICATimeoutNanos(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		return fmt.Errorf("Failed to get ICA timeout nanos for epochType %s using param, error: %s: %s", epochstypes.STRIDE_EPOCH, err.Error(), types.ErrInvalidRequest.Error())
	}

	// Submit the ICQ for the withdrawal account balance
	if err := k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_WithdrawalBalance,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "bank" store to access acct balances which live in the bank module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for withdrawal balance, error: %s", err.Error()))
		return err
	}

	return nil
}

// helper to get time at which next epoch begins, in unix nano units
func (k Keeper) GetStartTimeNextEpoch(ctx sdk.Context, epochType string) (uint64, error) {
	epochTracker, found := k.GetEpochTracker(ctx, epochType)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to get epoch tracker for %s", epochType))
		return 0, fmt.Errorf("Failed to get epoch tracker for %s: %s", epochType, types.ErrInvalidRequest.Error())
	}
	return epochTracker.NextEpochStartTime, nil
}

func (k Keeper) SubmitTxsDayEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
	account types.ICAAccount,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	sequence, err := k.SubmitTxsEpoch(ctx, connectionId, msgs, account, epochstypes.DAY_EPOCH, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

func (k Keeper) SubmitTxsStrideEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
	account types.ICAAccount,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	sequence, err := k.SubmitTxsEpoch(ctx, connectionId, msgs, account, epochstypes.STRIDE_EPOCH, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

func (k Keeper) SubmitTxsEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
	account types.ICAAccount,
	epochType string,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	timeoutNanosUint64, err := k.GetICATimeoutNanos(ctx, epochType)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to get ICA timeout nanos for epochType %s using param, error: %s", epochType, err.Error()))
		return 0, fmt.Errorf("Failed to convert timeoutNanos to uint64, error: %s: %s", err.Error(), types.ErrInvalidRequest.Error())
	}
	sequence, err := k.SubmitTxs(ctx, connectionId, msgs, account, timeoutNanosUint64, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

// SubmitTxs submits an ICA transaction containing multiple messages
func (k Keeper) SubmitTxs(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
	account types.ICAAccount,
	timeoutTimestamp uint64,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	chainId, err := k.GetChainID(ctx, connectionId)
	if err != nil {
		return 0, err
	}
	owner := types.FormatICAAccountOwner(chainId, account.Target)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return 0, err
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  Submitting ICA Tx on %s, %s with TTL: %d", portID, connectionId, timeoutTimestamp))
	for _, msg := range msgs {
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "    Msg: %+v", msg))
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
	if !found {
		return 0, fmt.Errorf("failed to retrieve active channel for port %s: %s", portID, icatypes.ErrActiveChannelNotFound.Error())
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return 0, fmt.Errorf("module does not own channel capability: %s", channeltypes.ErrChannelCapabilityNotFound.Error())
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return 0, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	sequence, err := k.ICAControllerKeeper.SendTx(ctx, chanCap, connectionId, portID, packetData, timeoutTimestamp)
	if err != nil {
		return 0, err
	}

	// Store the callback data
	if callbackId != "" && callbackArgs != nil {
		callback := icacallbackstypes.CallbackData{
			CallbackKey:  icacallbackstypes.PacketID(portID, channelID, sequence),
			PortId:       portID,
			ChannelId:    channelID,
			Sequence:     sequence,
			CallbackId:   callbackId,
			CallbackArgs: callbackArgs,
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Storing callback data: %+v", callback))
		k.ICACallbacksKeeper.SetCallbackData(ctx, callback)
	}

	return sequence, nil
}

func (k Keeper) GetLightClientHeightSafely(ctx sdk.Context, connectionID string) (uint64, error) {
	// get light client's latest height
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		errMsg := fmt.Sprintf("invalid connection id, %s not found", connectionID)
		k.Logger(ctx).Error(errMsg)
		return 0, fmt.Errorf(errMsg)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		errMsg := fmt.Sprintf("client id %s not found for connection %s", conn.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return 0, fmt.Errorf(errMsg)
	} else {
		latestHeightHostZone, err := cast.ToUint64E(clientState.GetLatestHeight().GetRevisionHeight())
		if err != nil {
			errMsg := fmt.Sprintf("error casting latest height to int64: %s", err.Error())
			k.Logger(ctx).Error(errMsg)
			return 0, fmt.Errorf(errMsg)
		}
		return latestHeightHostZone, nil
	}
}

func (k Keeper) GetLightClientTimeSafely(ctx sdk.Context, connectionID string) (uint64, error) {
	// get light client's latest height
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		errMsg := fmt.Sprintf("invalid connection id, %s not found", connectionID)
		k.Logger(ctx).Error(errMsg)
		return 0, fmt.Errorf(errMsg)
	}
	// TODO(TEST-112) make sure to update host LCs here!
	latestConsensusClientState, found := k.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(ctx, conn.ClientId)
	if !found {
		errMsg := fmt.Sprintf("client id %s not found for connection %s", conn.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return 0, fmt.Errorf(errMsg)
	} else {
		latestTime := latestConsensusClientState.GetTimestamp()
		return latestTime, nil
	}
}

// Submits an ICQ to get a validator's exchange rate
func (k Keeper) QueryValidatorExchangeRate(ctx sdk.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	k.Logger(ctx).Info(utils.LogWithHostZone(msg.ChainId, "Submitting ICQ for validator exchange rate to %s", msg.Valoper))

	// Ensure ICQ can be issued now! else fail the callback
	withinBufferWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to determine if ICQ callback is inside buffer window, err: %s: %s", err.Error(), types.ErrInvalidRequest.Error())
	} else if !withinBufferWindow {
		return nil, fmt.Errorf("outside the buffer time during which ICQs are allowed (%s): %s", msg.ChainId, types.ErrOutsideIcqWindow.Error())
	}

	// Confirm the host zone exists
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, fmt.Errorf("Host zone not found (%s): %s", msg.ChainId, types.ErrInvalidHostZone.Error())
	}

	// check that the validator address matches the bech32 prefix of the hz
	if !strings.Contains(msg.Valoper, hostZone.Bech32Prefix) {
		return nil, fmt.Errorf("validator operator address must match the host zone bech32 prefix: %s", types.ErrInvalidRequest.Error())
	}

	// Encode the validator address to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(msg.Valoper)
	if err != nil {
		return nil, fmt.Errorf("invalid validator operator address, could not decode (%s): %s", err.Error(), types.ErrInvalidRequest.Error())
	}
	queryData := stakingtypes.GetValidatorKey(validatorAddressBz)

	// The query should timeout at the start of the next epoch
	ttl, err := k.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		return nil, fmt.Errorf("could not get start time for next epoch: %s: %s", err.Error(), types.ErrInvalidRequest.Error())
	}

	// Submit validator exchange rate ICQ
	if err := k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_Validator,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "staking" store to access validator which lives in the staking module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for validator exchange rate, error %s", err.Error()))
		return nil, err
	}
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}

// Submits an ICQ to get a validator's delegations
// This is called after the validator's exchange rate is determined
func (k Keeper) QueryDelegationsIcq(ctx sdk.Context, hostZone types.HostZone, valoper string) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for delegations to %s", valoper))

	// Ensure ICQ can be issued now! else fail the callback
	valid, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine if ICQ callback is inside buffer window, err: %s: %s", err.Error(), types.ErrInvalidRequest.Error())
	} else if !valid {
		return fmt.Errorf("outside the buffer time during which ICQs are allowed (%s): %s", hostZone.HostDenom, types.ErrOutsideIcqWindow.Error())
	}

	// Get the validator and delegator encoded addresses to form the query request
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return fmt.Errorf("no delegation address found for %s: %s", hostZone.ChainId, types.ErrICAAccountNotFound.Error())
	}
	_, validatorAddressBz, err := bech32.DecodeAndConvert(valoper)
	if err != nil {
		return fmt.Errorf("invalid validator address, could not decode (%s): %s", err.Error(), types.ErrInvalidRequest.Error())
	}
	_, delegatorAddressBz, err := bech32.DecodeAndConvert(delegationAccount.Address)
	if err != nil {
		return fmt.Errorf("invalid delegator address, could not decode (%s): %s", err.Error(), types.ErrInvalidRequest.Error())
	}
	queryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)

	// The query should timeout at the start of the next epoch
	ttl, err := k.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		return fmt.Errorf("could not get start time for next epoch: %s: %s", err.Error(), types.ErrInvalidRequest.Error())
	}

	// Submit delegator shares ICQ
	if err := k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_Delegation,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "staking" store to access delegation which lives in the staking module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return err
	}

	return nil
}
