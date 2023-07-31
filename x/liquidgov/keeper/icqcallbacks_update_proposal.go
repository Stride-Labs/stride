package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v11/utils"
	icqtypes "github.com/Stride-Labs/stride/v11/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

func UpdateProposalCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_UpdateProposals,
		"Starting update proposals callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	_, found := k.stakeibcKeeper.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(stakeibctypes.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	queriedProposal := govtypes.Proposal{}
	err := k.cdc.Unmarshal(args, &queriedProposal)
	if err != nil {
		return sdkerrors.Wrapf(stakeibctypes.ErrMarshalFailure, "unable to unmarshal query response into Delegation type, err: %s", err.Error())
	}

	if queriedProposal.Status == govtypes.StatusVotingPeriod {
		liquidProp, _ := types.NewProposal(chainId, &queriedProposal)
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals, "Setting new proposal %+v", liquidProp))
		k.SetProposal(ctx, *liquidProp)
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals,
			"Added proposal #%d from host zone: %s VoteStartTime: %s VoteEndTime: %s", liquidProp.GovProposal.ProposalId, liquidProp.HostZoneId, liquidProp.GovProposal.VotingStartTime, liquidProp.GovProposal.VotingEndTime))
	} else {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals,
			"Proposal #%d from host zone: %s not in voting period", queriedProposal.ProposalId, chainId))
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_UpdateProposals, "Query response - Chain ID: %s Proposal ID: %d, VotingStartTime: %s, VotingEndTime: %s",
		chainId, queriedProposal.ProposalId, queriedProposal.VotingStartTime, queriedProposal.VotingEndTime))

	return nil
}	
