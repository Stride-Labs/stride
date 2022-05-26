package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// SubmitTx submits an ICA transaction
// NOTE: this is not a standard message; only the stakeibc module can call this function
func (k Keeper) DelegateOnHost(goCtx context.Context, hostZone types.HostZone, amt sdk.Coin) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx
	var msgs []sdk.Msg
	// TODO(TEST-38): Do we need to map between ports and connections? If so we can push this down to SubmitTx
	// I was under the impression there was a single connection for each pair of connected chains
	// (for some reason quicksilver maps between the two, but interchain-account-demo does not - they pull it off of the msg)
	// How will we get the proper connectionId here?
	connectionId := "hardcoded-connection"

	// Fetch the relevant ICA
	delegationIca := hostZone.GetDelegationAccount()
	
	// Construct the transaction
	// TODO(TEST-39): Implement validator selection
	validator_address := "hardcoded-validator-address"
	msgs = append(msgs, &stakingTypes.MsgDelegate{DelegatorAddress: delegationIca.GetAddress(), ValidatorAddress: validator_address, Amount: amt})
	// Send the transaction through SubmitTxs
	k.SubmitTxs(goCtx, connectionId, msgs, *delegationIca)

	return nil
}


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

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
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
	_, err = k.icaControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitTxResponse{}, nil
}


// RegisterAccount registers an ICA account on behalf of the stakeibc module
// NOTE: this is not a standard message; only the stakeibc module should call this function. However,
// this is temporarily in the message server to facilitate easy testing and development.
// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func (k Keeper) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.Owner); err != nil {
		return nil, err
	}
	// construct and return the ICAAccount, so that it can be stored

	// Return ICAAccount, err
	return &types.MsgRegisterAccountResponse{}, nil
}



// SubmitTx submits an ICA transaction
// NOTE: this is not a standard message; only the stakeibc module can call this function
func (k Keeper) SubmitTxs(goCtx context.Context, connectionId string, msgs []sdk.Msg, account types.ICAAccount) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Each ICA is associated with a port; we can dynamically reconstruct the port from the fields on ICA,
	// e.g. icacontroller-atom-{module_address}-DELEGATION
	// NOTE: ports are scoped to chains
	// How can we fetch chainId here?
	chainId := "hardcoded-chainId"
	// Sanity check the output of account.GetTarget() (not sure if it prints an int or a stringified ICAAccountType)
	owner := fmt.Sprintf("%s-%s-%s", account.GetAddress(), chainId, account.GetTarget().String())
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
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
	timeoutTimestamp := time.Now().Add(time.Minute).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, chanCap, connectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return err
	}

	return nil
}


