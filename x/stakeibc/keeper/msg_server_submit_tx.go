package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// SubmitTx sends an ICA transaction to a host chain on behalf of an account on the controller
// chain.
// NOTE: this is not a standard message; only the stakeibc module should call this function. However,
// this is temporarily in the message server to facilitate easy testing and development.
// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func (k Keeper) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	portID, err := icatypes.NewControllerPortID(msg.Owner)
	if err != nil {
		return nil, err
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		return nil, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, []sdk.Msg{msg.GetTxMsg()})
	if err != nil {
		return nil, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	// timeoutTimestamp := time.Now().Add(time.Minute).UnixNano()
	timeoutTimestamp := ^uint64(0) >> 1
	_, err = k.ICAControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitTxResponse{}, nil
}

func (k Keeper) DelegateOnHost(ctx sdk.Context, hostZone types.HostZone, amt sdk.Coin) error {
	_ = ctx
	var msgs []sdk.Msg
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
	delegationIca := hostZone.GetDelegationAccount()

	// Construct the transaction
	// TODO(TEST-39): Implement validator selection
	validator_address := "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne" // gval2

	// construct the msg
	msgs = append(msgs, &stakingTypes.MsgDelegate{DelegatorAddress: delegationIca.GetAddress(), ValidatorAddress: validator_address, Amount: amt})
	// Send the transaction through SubmitTx
	err = k.SubmitTxs(ctx, connectionId, msgs, *delegationIca)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}
	return nil
}

func (k Keeper) SetWithdrawalAddressOnHost(ctx sdk.Context, hostZone types.HostZone) error {
	_ = ctx
	var msgs []sdk.Msg
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
	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.Address == "" {
		k.Logger(ctx).Error("Zone %s is missing a delegation address!", hostZone.ChainId)
		return nil
	}
	withdrawalIca := hostZone.GetWithdrawalAccount()
	if withdrawalIca == nil || withdrawalIca.Address == "" {
		k.Logger(ctx).Error("Zone %s is missing a withdrawal address!", hostZone.ChainId)
		return nil
	}
	withdrawalIcaAddr := hostZone.GetWithdrawalAccount().Address

	// construct the msg
	msgs = append(msgs, &distributiontypes.MsgSetWithdrawAddress{DelegatorAddress: delegationIca.GetAddress(), WithdrawAddress: withdrawalIcaAddr})
	// Send the transaction through SubmitTx
	err = k.SubmitTxs(ctx, connectionId, msgs, *delegationIca)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}
	return nil
}

// Simple balance query helper using new ICQ module
func (k Keeper) UpdateWithdrawalBalance(ctx sdk.Context, zoneInfo types.HostZone) {
	k.Logger(ctx).Info(fmt.Sprintf("\tUpdating withdrawal balances on %s", zoneInfo.ChainId))

	withdrawalIca := zoneInfo.GetWithdrawalAccount()
	if withdrawalIca == nil || withdrawalIca.Address == "" {
		k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
	}
	k.Logger(ctx).Info(fmt.Sprintf("\tQuerying withdrawalBalances for %s at %d height", zoneInfo.ChainId))

	_, addr, _ := bech32.DecodeAndConvert(withdrawalIca.GetAddress())
	data := bankTypes.CreateAccountBalancesPrefix(addr)
	key := "store/bank/key"
	k.Logger(ctx).Info("Querying for value", "key", key, "denom", zoneInfo.HostDenom)
	k.InterchainQueryKeeper.MakeRequest(
		ctx,
		zoneInfo.ConnectionId,
		zoneInfo.ChainId,
		key,
		append(data, []byte(zoneInfo.HostDenom)...),
		sdk.NewInt(-1),
		types.ModuleName,
		"withdrawalbalance",
		0, //ttl
		0, //height
	)
}

// Simple query helper to get the current clock time of the host chain
func (k Keeper) ReadClockTime(ctx sdk.Context, zoneInfo types.HostZone) {
	// k.Logger(ctx).Info(fmt.Sprintf("\tQuerying clock time on %s", zoneInfo.ChainId))

	// withdrawalIca := zoneInfo.GetWithdrawalAccount()
	// if withdrawalIca == nil || withdrawalIca.Address == "" {
	// 	k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
	// }
	// k.Logger(ctx).Info(fmt.Sprintf("\tQuerying withdrawalBalances for %s at %d height", zoneInfo.ChainId))

	// _, addr, _ := bech32.DecodeAndConvert(withdrawalIca.GetAddress())
	// data := bankTypes.CreateAccountBalancesPrefix(addr)
	// key := "store/bank/key"
	// k.Logger(ctx).Info("Querying for value", "key", key, "denom", zoneInfo.HostDenom)
	// k.InterchainQueryKeeper.MakeRequest(
	// 	ctx,
	// 	zoneInfo.ConnectionId,
	// 	zoneInfo.ChainId,
	// 	key,
	// 	append(data, []byte(zoneInfo.HostDenom)...),
	// 	sdk.NewInt(-1),
	// 	types.ModuleName,
	// 	"withdrawalbalance",
	// 	0, //ttl
	// 	0, //height
	// )
}

// SubmitTxs submits an ICA transaction containing multiple messages
func (k Keeper) SubmitTxs(ctx sdk.Context, connectionId string, msgs []sdk.Msg, account types.ICAAccount) error {
	chainId, err := k.GetChainID(ctx, connectionId)
	if err != nil {
		return err
	}
	owner := types.FormatICAAccountOwner(chainId, account.GetTarget())
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
	if !found {
		return sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	// TODO(TEST-37): Decide on timeout logic
	timeoutTimestamp := ^uint64(0) >> 1
	_, err = k.ICAControllerKeeper.SendTx(ctx, chanCap, connectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) GetLightClientHeightSafely(ctx sdk.Context, connectionID string) (int64, bool) {

	var latestHeightHostZone int64 // defaults to 0
	// get light client's latest height
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("invalid connection id, \"%s\" not found", connectionID))
	}
	//TODO(TEST-112) make sure to update host LCs here!
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("client id \"%s\" not found for connection \"%s\"", conn.ClientId, connectionID))
		return 0, false
	} else {
		// TODO(TEST-119) get stAsset supply at SAME time as hostZone height
		// TODO(TEST-112) check on safety of castng uint64 to int64
		latestHeightHostZone = int64(clientState.GetLatestHeight().GetRevisionHeight())
		return latestHeightHostZone, true
	}
}

func (k Keeper) GetLightClientTimeSafely(ctx sdk.Context, connectionID string) (uint64, bool) {

	// get light client's latest height
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("invalid connection id, \"%s\" not found", connectionID))
	}
	//TODO(TEST-112) make sure to update host LCs here!
	latestConsensusClientState, found := k.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(ctx, conn.ClientId)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("client id \"%s\" not found for connection \"%s\"", conn.ClientId, connectionID))
		return 0, false
	} else {
		latestTime := latestConsensusClientState.GetTimestamp()
		return latestTime, true
	}
}
