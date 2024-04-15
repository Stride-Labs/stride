package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/gogoproto/proto"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"

	"github.com/Stride-Labs/stride/v21/utils"
	epochstypes "github.com/Stride-Labs/stride/v21/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v21/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

const (
	ClaimRewardsICABatchSize = 10
)

func (k Keeper) SetWithdrawalAddressOnHost(ctx sdk.Context, hostZone types.HostZone) error {
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
	_, err := k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, types.ICAAccountType_DELEGATION, "", nil)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", hostZone.ConnectionId, hostZone.ChainId, msgs)
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

// Registers a new TradeRoute ICAAccount, given the type
// Stores down the connection and chainId now, and the address upon callback
func (k Keeper) RegisterTradeRouteICAAccount(
	ctx sdk.Context,
	tradeRouteId string,
	connectionId string,
	icaAccountType types.ICAAccountType,
) (account types.ICAAccount, err error) {
	// Get the chain ID and counterparty connection-id from the connection ID on Stride
	chainId, err := k.GetChainIdFromConnectionId(ctx, connectionId)
	if err != nil {
		return account, err
	}
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionId)
	if !found {
		return account, errorsmod.Wrap(connectiontypes.ErrConnectionNotFound, connectionId)
	}
	counterpartyConnectionId := connection.Counterparty.ConnectionId

	// Build the appVersion, owner, and portId needed for registration
	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: connectionId,
		HostConnectionId:       counterpartyConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
	owner := types.FormatTradeRouteICAOwnerFromRouteId(chainId, tradeRouteId, icaAccountType)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return account, err
	}

	// Create the associate ICAAccount object
	account = types.ICAAccount{
		ChainId:      chainId,
		Type:         icaAccountType,
		ConnectionId: connectionId,
	}

	// Check if an ICA account has already been created
	// (in the event that this trade route was removed and then added back)
	// If so, there's no need to register a new ICA
	_, channelFound := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, connectionId, portID)
	icaAddress, icaFound := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, connectionId, portID)
	if channelFound && icaFound {
		account = types.ICAAccount{
			ChainId:      chainId,
			Type:         icaAccountType,
			ConnectionId: connectionId,
			Address:      icaAddress,
		}
		return account, nil
	}

	// Otherwise, if there's no account already, register a new one
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, connectionId, owner, appVersion); err != nil {
		return account, err
	}

	return account, nil
}
