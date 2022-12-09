package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

func (k Keeper) DelegateOnHost(ctx sdk.Context, hostZone types.HostZone, amt sdk.Coin, depositRecord recordstypes.DepositRecord) error {
	// the relevant ICA is the delegate account
	owner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidChainID, "%s has no associated connection", portID)
	}

	// Fetch the relevant ICA
	delegationIca := hostZone.DelegationAccount
	if delegationIca == nil || delegationIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegation account")
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
				DelegatorAddress: delegationIca.Address,
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
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationIca, ICACallbackID_Delegate, marshalledCallbackArgs)
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to SubmitTxs for connectionId %s on %s. Messages: %s", connectionId, hostZone.ChainId, msgs)
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidChainID, "%s has no associated connection", portID)
	}

	// Fetch the relevant ICA
	delegationIca := hostZone.DelegationAccount
	if delegationIca == nil || delegationIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return nil
	}
	withdrawalIca := hostZone.WithdrawalAccount
	if withdrawalIca == nil || withdrawalIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a withdrawal address!", hostZone.ChainId))
		return nil
	}
	withdrawalIcaAddr := hostZone.WithdrawalAccount.Address

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Withdrawal Address: %s, Delegator Address: %s", withdrawalIcaAddr, delegationIca.Address))

	// Construct the ICA message
	msgs := []sdk.Msg{
		&distributiontypes.MsgSetWithdrawAddress{
			DelegatorAddress: delegationIca.Address,
			WithdrawAddress:  withdrawalIcaAddr,
		},
	}
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationIca, "", nil)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}

	return nil
}

// Simple balance query helper using new ICQ module
func (k Keeper) UpdateWithdrawalBalance(ctx sdk.Context, hostZone types.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for withdrawal account balance"))

	withdrawalIca := hostZone.WithdrawalAccount
	if withdrawalIca == nil || withdrawalIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a withdrawal address!", hostZone.ChainId))
	}

	_, addr, _ := bech32.DecodeAndConvert(withdrawalIca.Address)
	data := bankTypes.CreateAccountBalancesPrefix(addr)

	// get ttl, the end of the ICA buffer window
	epochType := epochstypes.STRIDE_EPOCH
	ttl, err := k.GetICATimeoutNanos(ctx, epochType)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get ICA timeout nanos for epochType %s using param, error: %s", epochType, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_WithdrawalBalance,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "bank" store to access acct balances which live in the bank module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		append(data, []byte(hostZone.HostDenom)...),
		ttl, // ttl
	)
	if err != nil {
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
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to get epoch tracker for %s", epochType)
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
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to convert timeoutNanos to uint64, error: %s", err.Error())
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
		return 0, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
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

// query and update validator exchange rate
func (k Keeper) QueryValidatorExchangeRate(ctx sdk.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	// ensure ICQ can be issued now! else fail the callback
	valid, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return nil, err
	} else if !valid {
		return nil, sdkerrors.Wrapf(types.ErrOutsideIcqWindow, "outside the buffer time during which ICQs are allowed (%s)", msg.ChainId)
	}

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		errMsg := fmt.Sprintf("Host zone not found (%s)", msg.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, errMsg)
	}

	// check that the validator address matches the bech32 prefix of the hz
	if !strings.Contains(msg.Valoper, hostZone.Bech32Prefix) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must match the host zone bech32 prefix")
	}

	_, valAddr, err := bech32.DecodeAndConvert(msg.Valoper)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator operator address, could not decode (%s)", err.Error())
	}
	data := stakingtypes.GetValidatorKey(valAddr)

	// get ttl
	ttl, err := k.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		errMsg := fmt.Sprintf("could not get start time for next epoch: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Querying validator %v, key %v, denom %v", msg.Valoper, icqtypes.STAKING_STORE_QUERY_WITH_PROOF, hostZone.ChainId))
	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_Validator,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "staking" store to access validator which lives in the staking module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		data,
		ttl, // ttl
	)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for validator, error %s", err.Error()))
		return nil, err
	}
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}

// to icq delegation amounts, this fn is executed after validator exch rates are icq'd
func (k Keeper) QueryDelegationsIcq(ctx sdk.Context, hostZone types.HostZone, valoper string) error {
	// ensure ICQ can be issued now! else fail the callback
	valid, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return err
	} else if !valid {
		return sdkerrors.Wrapf(types.ErrOutsideIcqWindow, "outside the buffer time during which ICQs are allowed (%s)", hostZone.HostDenom)
	}

	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.GetAddress() == "" {
		errMsg := fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}
	delegationAcctAddr := delegationIca.GetAddress()
	_, valAddr, _ := bech32.DecodeAndConvert(valoper)
	_, delAddr, _ := bech32.DecodeAndConvert(delegationAcctAddr)
	data := stakingtypes.GetDelegationKey(delAddr, valAddr)

	// get ttl
	ttl, err := k.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		errMsg := fmt.Sprintf("could not get start time for next epoch: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Querying delegation for %s on %s", delegationAcctAddr, valoper))
	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_Delegation,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "staking" store to access delegation which lives in the staking module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		data,
		ttl, // ttl
	)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for delegation, error : %s", err.Error()))
		return err
	}
	return nil
}
