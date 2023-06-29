package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v10/utils"
	icqtypes "github.com/Stride-Labs/stride/v10/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v10/x/stakeibc/types"
)

func UpdateProposalCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_UpdateProposals,
		"Starting mirror proposals callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	_, found := k.stakeibcKeeper.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(stakeibctypes.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	highestID, _ := k.GetProposalID(ctx, chainId)

	queriedProposal := govtypesv1beta1.Proposal{}

	err := k.cdc.Unmarshal(args, &queriedProposal)
	if err != nil {
		return sdkerrors.Wrapf(stakeibctypes.ErrMarshalFailure, "unable to unmarshal query response into Delegation type, err: %s", err.Error())
	}

	if queriedProposal.ProposalId > highestID {
		if queriedProposal.Status == govtypesv1beta1.StatusVotingPeriod {
			liquidProp := types.Proposal{HostZoneId: chainId, GovProposal: queriedProposal}
			k.SetProposal(ctx, liquidProp)
			k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals,
				"Added proposal #%d from host zone: %s VoteStartTime: %s VoteEndTime: %s", liquidProp.GovProposal.ProposalId, liquidProp.HostZoneId, liquidProp.GovProposal.VotingStartTime, liquidProp.GovProposal.VotingEndTime))
		} else {
			k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals,
				"Proposal #%d from host zone: %s not in voting period. Incrementing highest ID.", queriedProposal.ProposalId, chainId))
		}
		k.SetProposalID(ctx, highestID+1, chainId)

	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals, "Query response - Chain ID: %s Proposal ID: %d, VotingStartTime: %s, VotingEndTime: %s",
		chainId, queriedProposal.ProposalId, queriedProposal.VotingStartTime, queriedProposal.VotingEndTime))

	return nil
}	
