package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

func (k Keeper) MirrorProposals(ctx sdk.Context, hostZone stakeibctypes.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for proposals to %s", hostZone.ChainId))

	// get highest proposal ID on stride
	highestID, _ := k.GetProposalID(ctx, hostZone.ChainId)

	// query for next proposal ID
	queryData := govtypes.ProposalKey(highestID + 1)

	// The query should timeout at the start of the next epoch
	ttl, err := k.stakeibcKeeper.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not get start time for next epoch: %s", err.Error())
	}

	if err := k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_MirrorProposals,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "gov" store to access proposals which live in the gov module
		// use "key" suffix to retrieve proposals by key
		icqtypes.GOV_STORE_QUERY_KEY,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for proposals, error : %s", err.Error()))
		return err
	}
	return nil
}

func (k Keeper) CastVoteOnHost(ctx sdk.Context, hostZone stakeibctypes.HostZone, govVote govtypesv1beta1.Vote) error {
	// the relevant ICA is the delegate account
	owner := stakeibctypes.FormatICAAccountOwner(hostZone.ChainId, stakeibctypes.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, err := k.stakeibcKeeper.GetConnectionId(ctx, portID)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidChainID, "%s has no associated connection", portID)
	}

	// Fetch the relevant ICA
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegation account")
	}

	var msgs []sdk.Msg
	msgs = append(msgs, govtypesv1beta1.NewMsgVote(
		sdk.AccAddress(govVote.Voter), govVote.ProposalId, govVote.Options[0].Option))

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Preparing MsgVote from the delegation account"))

	// add callback data
	castVoteOnHostCallback := types.CastVoteOnHostCallback{
		HostZoneId: hostZone.ChainId,
		GovVote:    govVote,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Marshalling CastVoteOnHostCallback args: %+v", castVoteOnHostCallback))
	marshalledCallbackArgs, err := k.MarshalCastVoteOnHostCallbackArgs(ctx, castVoteOnHostCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationAccount, ICACallbackID_CastVoteOnHost, marshalledCallbackArgs)
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to SubmitTxs for connectionId %s on %s. Messages: %s", connectionId, hostZone.ChainId, msgs)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgVote Successfully Sent"))

	return nil
}

func (k Keeper) SubmitTxsStrideEpoch(
	ctx sdk.Context,
	connectionId string,
	msgs []sdk.Msg,
	account stakeibctypes.ICAAccount,
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
	account stakeibctypes.ICAAccount,
	epochType string,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	timeoutNanosUint64, err := k.stakeibcKeeper.GetICATimeoutNanos(ctx, epochType)
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
	account stakeibctypes.ICAAccount,
	timeoutTimestamp uint64,
	callbackId string,
	callbackArgs []byte,
) (uint64, error) {
	chainId, err := k.stakeibcKeeper.GetChainID(ctx, connectionId)
	if err != nil {
		return 0, err
	}
	owner := stakeibctypes.FormatICAAccountOwner(chainId, account.Target)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return 0, err
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  Submitting ICA Tx on %s, %s with TTL: %d", portID, connectionId, timeoutTimestamp))
	for _, msg := range msgs {
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "    Msg: %+v", msg))
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  166"))

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
	if !found {
		return 0, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  172"))

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  175 %s %v %s", types.ModuleName, chanCap, found))

	if !found {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  178"))

	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return 0, err
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  184"))

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  190"))

	sequence, err := k.ICAControllerKeeper.SendTx(ctx, chanCap, connectionId, portID, packetData, timeoutTimestamp)
	if err != nil {
		return 0, err
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  196"))

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
