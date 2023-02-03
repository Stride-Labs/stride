package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"

	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

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

func (k Keeper) CastVoteOnHost(ctx sdk.Context, hostZone stakeibctypes.HostZone, govVote govtypesv1.Vote) error {
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
	msgs = append(msgs, govtypesv1.NewMsgVote(
		sdk.AccAddress(govVote.Voter), govVote.ProposalId, govVote.Options[0].Option, ""))

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
	_, err = k.stakeibcKeeper.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *delegationAccount, ICACallbackID_CastVoteOnHost, marshalledCallbackArgs)
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to SubmitTxs for connectionId %s on %s. Messages: %s", connectionId, hostZone.ChainId, msgs)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgVote Successfully Sent"))

	return nil
}
