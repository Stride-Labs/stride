package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cast"

	icacallbackstypes "github.com/Stride-Labs/stride/v3/x/icacallbacks/types"

	recordstypes "github.com/Stride-Labs/stride/v3/x/records/types"
	"github.com/Stride-Labs/stride/v3/x/stakeibc/types"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v3/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v3/x/interchainquery/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

func (k Keeper) DelegateOnHost(ctx sdk.Context, hostZone types.HostZone, amt sdk.Coin, depositRecord recordstypes.DepositRecord) error {
	var msgs []sdk.Msg

	// the relevant ICA is the delegate account
	owner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return fmt.Errorf("%s has no associated portId: invalid address", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return fmt.Errorf("%s has no associated connection: invalid chain-id", portID)
	}

	// Fetch the relevant ICA
	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.GetAddress() == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return fmt.Errorf("Invalid delegation account: : invalid address")
	}

	// Construct the transaction
	targetDelegatedAmts, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, amt.Amount.Uint64())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target delegation amounts for host zone %s", hostZone.ChainId))
		return err
	}

	var splitDelegations []*types.SplitDelegation
	for _, validator := range hostZone.GetValidators() {
		relativeAmount := sdk.NewCoin(amt.Denom, sdk.NewIntFromUint64(targetDelegatedAmts[validator.GetAddress()]))
		if relativeAmount.Amount.IsPositive() {
			k.Logger(ctx).Info(fmt.Sprintf("Appending MsgDelegate to msgs, DelegatorAddress: %s, ValidatorAddress: %s, relativeAmount: %v",
				delegationIca.GetAddress(), validator.GetAddress(), relativeAmount))

			msgs = append(msgs, &stakingTypes.MsgDelegate{
				DelegatorAddress: delegationIca.GetAddress(),
				ValidatorAddress: validator.GetAddress(),
				Amount:           relativeAmount,
			})
		}
		splitDelegations = append(splitDelegations, &types.SplitDelegation{
			Validator: validator.GetAddress(),
			Amount:    relativeAmount.Amount.Uint64(),
		})
	}

	// add callback data
	delegateCallback := types.DelegateCallback{
		HostZoneId:       hostZone.ChainId,
		DepositRecordId:  depositRecord.Id,
		SplitDelegations: splitDelegations,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Marshalling DelegateCallback args: %v", delegateCallback))
	marshalledCallbackArgs, err := k.MarshalDelegateCallbackArgs(ctx, delegateCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationIca, ICACallbackID_Delegate, marshalledCallbackArgs)
	if err != nil {
		return fmt.Errorf("Failed to SubmitTxs for connectionId %s on %s. Messages: %s: %s", connectionId, hostZone.ChainId, msgs, err.Error())
	}
	// update the record state to DELEGATION_IN_PROGRESS
	depositRecord.Status = recordstypes.DepositRecord_DELEGATION_IN_PROGRESS
	k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
	return nil
}

func (k Keeper) SetWithdrawalAddressOnHost(ctx sdk.Context, hostZone types.HostZone) error {
	_ = ctx
	var msgs []sdk.Msg
	// the relevant ICA is the delegate account
	owner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return fmt.Errorf("%s has no associated portId: invalid address", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return fmt.Errorf("%s has no associated connection: invalid chain-id", portID)
	}

	// Fetch the relevant ICA
	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return nil
	}
	withdrawalIca := hostZone.GetWithdrawalAccount()
	if withdrawalIca == nil || withdrawalIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a withdrawal address!", hostZone.ChainId))
		return nil
	}
	withdrawalIcaAddr := hostZone.GetWithdrawalAccount().GetAddress()

	k.Logger(ctx).Info(fmt.Sprintf("Setting withdrawal address on host zone. DelegatorAddress: %s WithdrawAddress: %s ConnectionID: %s", delegationIca.GetAddress(), withdrawalIcaAddr, connectionId))
	// construct the msg
	msgs = append(msgs, &distributiontypes.MsgSetWithdrawAddress{DelegatorAddress: delegationIca.GetAddress(), WithdrawAddress: withdrawalIcaAddr})
	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationIca, "", nil)
	if err != nil {
		return fmt.Errorf("Failed to SubmitTxs for %s, %s, %s: invalid request", connectionId, hostZone.ChainId, msgs)
	}
	return nil
}

// Simple balance query helper using new ICQ module
func (k Keeper) UpdateWithdrawalBalance(ctx sdk.Context, zoneInfo types.HostZone) error {
	k.Logger(ctx).Info(fmt.Sprintf("\tUpdating withdrawal balances on %s", zoneInfo.ChainId))

	withdrawalIca := zoneInfo.GetWithdrawalAccount()
	if withdrawalIca == nil || withdrawalIca.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a withdrawal address!", zoneInfo.ChainId))
	}
	k.Logger(ctx).Info(fmt.Sprintf("\tQuerying withdrawalBalances for %s", zoneInfo.ChainId))

	_, addr, _ := bech32.DecodeAndConvert(withdrawalIca.GetAddress())
	data := bankTypes.CreateAccountBalancesPrefix(addr)

	// get ttl, the end of the ICA buffer window
	epochType := epochstypes.STRIDE_EPOCH
	ttl, err := k.GetICATimeoutNanos(ctx, epochType)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get ICA timeout nanos for epochType %s using param, error: %s", epochType, err.Error())
		k.Logger(ctx).Error(errMsg)
		return fmt.Errorf("%s: %s", errMsg, "invalid request")
	}

	k.Logger(ctx).Info("Querying for value", "key", icqtypes.BANK_STORE_QUERY_WITH_PROOF, "denom", zoneInfo.HostDenom)
	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		zoneInfo.ConnectionId,
		zoneInfo.ChainId,
		// use "bank" store to access acct balances which live in the bank module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		append(data, []byte(zoneInfo.HostDenom)...),
		sdk.NewInt(-1),
		types.ModuleName,
		ICQCallbackID_WithdrawalBalance,
		ttl, // ttl
		0,   // height always 0 (which means current height)
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
		return 0, fmt.Errorf("Failed to get epoch tracker for %s: invalid request", epochType)
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
	k.Logger(ctx).Info(fmt.Sprintf("SubmitTxsDayEpoch %v", msgs))
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
	k.Logger(ctx).Info(fmt.Sprintf("SubmitTxsStrideEpoch %v", msgs))
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
		return 0, fmt.Errorf("Failed to convert timeoutNanos to uint64, error: %s: invalid request", err.Error())
	}
	sequence, err := k.SubmitTxs(ctx, connectionId, msgs, account, timeoutNanosUint64, callbackId, callbackArgs)
	if err != nil {
		return 0, err
	}
	k.Logger(ctx).Info(fmt.Sprintf("Submitted Txs, connectionId: %s, sequence: %d, block: %d", connectionId, sequence, ctx.BlockHeight()))
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
	k.Logger(ctx).Info(fmt.Sprintf("SubmitTxs %v", msgs))
	chainId, err := k.GetChainID(ctx, connectionId)
	if err != nil {
		return 0, err
	}
	owner := types.FormatICAAccountOwner(chainId, account.GetTarget())
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return 0, err
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
		return nil, fmt.Errorf("outside the buffer time during which ICQs are allowed (%s): %s", msg.ChainId, types.ErrOutsideIcqWindow.Error())
	}

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		errMsg := fmt.Sprintf("Host zone not found (%s)", msg.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, fmt.Errorf("%s: %s", errMsg, types.ErrInvalidHostZone.Error())
	}

	// check that the validator address matches the bech32 prefix of the hz
	if !strings.Contains(msg.Valoper, hostZone.Bech32Prefix) {
		return nil, fmt.Errorf("validator operator address must match the host zone bech32 prefix: invalid request")
	}

	_, valAddr, err := bech32.DecodeAndConvert(msg.Valoper)
	if err != nil {
		return nil, fmt.Errorf("invalid validator operator address, could not decode (%s): invalid request", err.Error())
	}
	data := stakingtypes.GetValidatorKey(valAddr)

	// get ttl
	ttl, err := k.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		errMsg := fmt.Sprintf("could not get start time for next epoch: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, fmt.Errorf("%s: %s", errMsg, "invalid request")
	}

	k.Logger(ctx).Info(fmt.Sprintf("Querying validator %v, key %v, denom %v", msg.Valoper, icqtypes.STAKING_STORE_QUERY_WITH_PROOF, hostZone.ChainId))
	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		hostZone.ConnectionId,
		hostZone.ChainId,
		// use "staking" store to access validator which lives in the staking module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		data,
		sdk.NewInt(-1),
		types.ModuleName,
		ICQCallbackID_Validator,
		ttl, // ttl
		0,   // height always 0 (which means current height)
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
		return fmt.Errorf("outside the buffer time during which ICQs are allowed (%s): %s", hostZone.HostDenom, types.ErrOutsideIcqWindow.Error())
	}

	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.GetAddress() == "" {
		errMsg := fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return fmt.Errorf("%s: %s", errMsg, types.ErrICAAccountNotFound.Error())
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
		return fmt.Errorf("%s: %s", errMsg, "invalid request")
	}

	k.Logger(ctx).Info(fmt.Sprintf("Querying delegation for %s on %s", delegationAcctAddr, valoper))
	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		hostZone.ConnectionId,
		hostZone.ChainId,
		// use "staking" store to access delegation which lives in the staking module
		// use "key" suffix to retrieve a proof alongside the query result
		icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		data,
		sdk.NewInt(-1),
		types.ModuleName,
		ICQCallbackID_Delegation,
		ttl, // ttl
		0,   // height always 0 (which means current height)
	)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for delegation, error : %s", err.Error()))
		return err
	}
	return nil
}
