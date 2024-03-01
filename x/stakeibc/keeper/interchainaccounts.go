package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v18/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v18/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v18/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v18/x/interchainquery/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

const (
	ClaimRewardsICABatchSize = 10
)

func (k Keeper) SetWithdrawalAddressOnHost(ctx sdk.Context, hostZone types.HostZone) error {
	// TODO: Remove this block and use connection-id from host zone
	// The relevant ICA is the delegate account
	owner := types.FormatHostZoneICAOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, found := k.GetConnectionIdFromICAPortId(ctx, portID)
	if !found {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "unable to find ICA connection Id for port %s", portID)
	}

	// Fetch the relevant ICA
	if hostZone.DelegationIcaAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return nil
	}

	if hostZone.WithdrawalIcaAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a withdrawal address!", hostZone.ChainId))
		return nil
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Withdrawal Address: %s, Delegator Address: %s",
		hostZone.WithdrawalIcaAddress, hostZone.DelegationIcaAddress))

	// Construct the ICA message
	msgs := []proto.Message{
		&distributiontypes.MsgSetWithdrawAddress{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			WithdrawAddress:  hostZone.WithdrawalIcaAddress,
		},
	}
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, types.ICAAccountType_DELEGATION, "", nil)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}

	return nil
}

func (k Keeper) ClaimAccruedStakingRewardsOnHost(ctx sdk.Context, hostZone types.HostZone) error {
	// Fetch the relevant ICA
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "delegation ICA not found for %s", hostZone.ChainId)
	}
	if hostZone.WithdrawalIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "withdrawal ICA not found for %s", hostZone.ChainId)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Withdrawal Address: %s, Delegator Address: %s",
		hostZone.WithdrawalIcaAddress, hostZone.DelegationIcaAddress))

	validators := hostZone.Validators

	// Build multi-message transaction to withdraw rewards from each validator
	// batching txs into groups of ClaimRewardsICABatchSize messages, to ensure they will fit in the host's blockSize
	for start := 0; start < len(validators); start += ClaimRewardsICABatchSize {
		end := start + ClaimRewardsICABatchSize
		if end > len(validators) {
			end = len(validators)
		}
		batch := validators[start:end]
		msgs := []proto.Message{}
		// Iterate over the items within the batch
		for _, val := range batch {
			// skip withdrawing rewards
			if val.Delegation.IsZero() {
				continue
			}
			msg := &distributiontypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: hostZone.DelegationIcaAddress,
				ValidatorAddress: val.Address,
			}
			msgs = append(msgs, msg)
		}

		if len(msgs) > 0 {
			_, err := k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, types.ICAAccountType_DELEGATION, "", nil)
			if err != nil {
				return errorsmod.Wrapf(err, "Failed to SubmitTxs for %s, %s, %s", hostZone.ConnectionId, hostZone.ChainId, msgs)
			}
		}
	}

	return nil
}

// Submits an ICQ for the withdrawal account balance
func (k Keeper) UpdateWithdrawalBalance(ctx sdk.Context, hostZone types.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for withdrawal account balance"))

	// Get the withdrawal account address from the host zone
	if hostZone.WithdrawalIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no withdrawal account found for %s", hostZone.ChainId)
	}

	// Encode the withdrawal account address for the query request
	// The query request consists of the withdrawal account address and denom
	_, withdrawalAddressBz, err := bech32.DecodeAndConvert(hostZone.WithdrawalIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid withdrawal account address, could not decode (%s)", err.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(withdrawalAddressBz), []byte(hostZone.HostDenom)...)

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_WithdrawalBalance,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for withdrawal balance, error: %s", err.Error()))
		return err
	}

	return nil
}

func (k Keeper) SubmitTxsDayEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []proto.Message,
	icaAccountType types.ICAAccountType,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	sequence, err := k.SubmitTxsEpoch(ctx, connectionId, msgs, icaAccountType, epochstypes.DAY_EPOCH, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

func (k Keeper) SubmitTxsStrideEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []proto.Message,
	icaAccountType types.ICAAccountType,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	sequence, err := k.SubmitTxsEpoch(ctx, connectionId, msgs, icaAccountType, epochstypes.STRIDE_EPOCH, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

func (k Keeper) SubmitTxsEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []proto.Message,
	icaAccountType types.ICAAccountType,
	epochType string,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	timeoutNanosUint64, err := k.GetICATimeoutNanos(ctx, epochType)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to get ICA timeout nanos for epochType %s using param, error: %s", epochType, err.Error()))
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to convert timeoutNanos to uint64, error: %s", err.Error())
	}
	sequence, err := k.SubmitTxs(ctx, connectionId, msgs, icaAccountType, timeoutNanosUint64, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

// SubmitTxs submits an ICA transaction containing multiple messages
// This function only supports messages to ICAs on the host zone
func (k Keeper) SubmitTxs(
	ctx sdk.Context,
	connectionId string,
	msgs []proto.Message,
	icaAccountType types.ICAAccountType,
	timeoutTimestamp uint64,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	chainId, err := k.GetChainIdFromConnectionId(ctx, connectionId)
	if err != nil {
		return 0, err
	}
	owner := types.FormatHostZoneICAOwner(chainId, icaAccountType)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return 0, err
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  Submitting ICA Tx on %s, %s with TTL: %d", portID, connectionId, timeoutTimestamp))
	protoMsgs := []proto.Message{}
	for _, msg := range msgs {
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "    Msg: %+v", msg))
		protoMsgs = append(protoMsgs, msg)
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
	if !found {
		return 0, errorsmod.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, protoMsgs)
	if err != nil {
		return 0, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// Submit ICA tx
	msgServer := icacontrollerkeeper.NewMsgServerImpl(&k.ICAControllerKeeper)
	relativeTimeoutOffset := timeoutTimestamp - uint64(ctx.BlockTime().UnixNano())
	msgSendTx := icacontrollertypes.NewMsgSendTx(owner, connectionId, relativeTimeoutOffset, packetData)
	res, err := msgServer.SendTx(ctx, msgSendTx)
	if err != nil {
		return 0, err
	}
	sequence := res.Sequence

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

func (k Keeper) SubmitICATxWithoutCallback(
	ctx sdk.Context,
	connectionId string,
	icaAccountOwner string,
	msgs []proto.Message,
	timeoutTimestamp uint64,
) error {
	// Serialize tx messages
	txBz, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to serialize cosmos transaction")
	}
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: txBz,
	}
	relativeTimeoutOffset := timeoutTimestamp - uint64(ctx.BlockTime().UnixNano())

	// Submit ICA, no need to store callback data or register callback function
	icaMsgServer := icacontrollerkeeper.NewMsgServerImpl(&k.ICAControllerKeeper)
	msgSendTx := icacontrollertypes.NewMsgSendTx(icaAccountOwner, connectionId, relativeTimeoutOffset, packetData)
	_, err = icaMsgServer.SendTx(ctx, msgSendTx)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to send ICA tx")
	}

	return nil
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

// Submits an ICQ to get a validator's delegations
// This is called after the validator's sharesToTokens rate is determined
// The timeoutDuration parameter represents the length of the timeout (not to be confused with an actual timestamp)
func (k Keeper) SubmitCalibrationICQ(ctx sdk.Context, hostZone types.HostZone, validatorAddress string) error {
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}

	// ensure the validator is in the set for this host
	_, _, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", validatorAddress)
	}

	// Get the validator and delegator encoded addresses to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator address, could not decode (%s)", err.Error())
	}
	_, delegatorAddressBz, err := bech32.DecodeAndConvert(hostZone.DelegationIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid delegator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)

	// Submit delegator shares ICQ
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Calibrate,
		CallbackData:    []byte{},
		TimeoutDuration: time.Hour,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return err
	}

	return nil
}
