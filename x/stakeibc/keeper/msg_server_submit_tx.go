package keeper

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto" //nolint:staticcheck
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"

	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	// TODO [LSM]: Revert type
	lsmdistributiontypes "github.com/iqlusioninc/liquidity-staking-module/x/distribution/types"
	lsmstakingtypes "github.com/iqlusioninc/liquidity-staking-module/x/staking/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"

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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidChainID, "%s has no associated connection", portID)
	}

	// Fetch the relevant ICA
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
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
			msgs = append(msgs, &lsmstakingtypes.MsgDelegate{
				DelegatorAddress: hostZone.DelegationIcaAddress,
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
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, types.ICAAccountType_DELEGATION, ICACallbackID_Delegate, marshalledCallbackArgs)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs for connectionId %s on %s. Messages: %s", connectionId, hostZone.ChainId, msgs)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgDelegates Successfully Sent"))

	// flag the delegation change in progress on each validator
	for _, splitDelegation := range splitDelegations {
		if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, splitDelegation.Validator); err != nil {
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)

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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidChainID, "%s has no associated connection", portID)
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
	msgs := []sdk.Msg{
		&lsmdistributiontypes.MsgSetWithdrawAddress{
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

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_WithdrawalBalance,
		TimeoutDuration: time.Hour,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
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
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to get epoch tracker for %s", epochType)
	}
	return epochTracker.NextEpochStartTime, nil
}

func (k Keeper) SubmitTxsDayEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
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
	msgs []sdk.Msg,
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
	msgs []sdk.Msg,
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
func (k Keeper) SubmitTxs(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
	icaAccountType types.ICAAccountType,
	timeoutTimestamp uint64,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	chainId, err := k.GetChainID(ctx, connectionId)
	if err != nil {
		return 0, err
	}
	owner := types.FormatICAAccountOwner(chainId, icaAccountType)
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
		return 0, errorsmod.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return 0, errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
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
func (k Keeper) QueryValidatorExchangeRate(
	ctx sdk.Context,
	chainId string,
	validatorAddress string,
	callbackDataBz []byte,
	aggressiveTimeout bool,
) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Submitting ICQ for validator exchange rate to %s", validatorAddress))

	// If this query is executed from LSM, there should be a more aggressive timeout since it's UX blocking,
	// and, in the event of a timeout, we should still enter the callback so we can alert the user that the query failed
	// If this query is not for an LSM liquid stake, we can have a more relaxed timeout
	var timeoutDuration time.Duration
	var timeoutPolicy icqtypes.TimeoutPolicy
	if aggressiveTimeout {
		timeoutDuration = LSMSlashQueryTimeout
		timeoutPolicy = icqtypes.TimeoutPolicy_EXECUTE_QUERY_CALLBACK
	} else {
		timeoutDuration = time.Hour
		timeoutPolicy = icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST
	}

	// Confirm the host zone exists
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidHostZone, "Host zone not found (%s)", chainId)
	}

	// check that the validator address matches the bech32 prefix of the hz
	if !strings.Contains(validatorAddress, hostZone.Bech32Prefix) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must match the host zone bech32 prefix")
	}

	// Encode the validator address to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator operator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetValidatorKey(validatorAddressBz)

	// Submit validator exchange rate ICQ
	// Considering this query is executed manually, we can be conservative with the timeout
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Validator,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   timeoutPolicy,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for validator exchange rate, error %s", err.Error()))
		return err
	}
	return nil
}

// Submits an ICQ to get a validator's delegations
// This is called after the validator's exchange rate is determined
// The timeoutDuration parameter represents the length of the timeout (not to be confused with an actual timestamp)
func (k Keeper) QueryDelegationsIcq(ctx sdk.Context, hostZone types.HostZone, validatorAddress string, timeoutDuration time.Duration) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for delegations to %s", validatorAddress))

	// Get the validator and delegator encoded addresses to form the query request
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", validatorAddress)
	}
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator address, could not decode (%s)", err.Error())
	}
	_, delegatorAddressBz, err := bech32.DecodeAndConvert(hostZone.DelegationIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid delegator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)

	// Store the current validator's delegation in the callback data so we can determine if it changed
	// while the query was in flight
	callbackData := types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: validator.Delegation,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal delegator shares callback data")
	}

	// Update the validator to indicate that the slash query is in progress
	validator.SlashQueryInProgress = true
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	// Submit delegator shares ICQ
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Delegation,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return err
	}

	return nil
}
