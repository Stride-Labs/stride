package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/Stride-Labs/stride/v5/utils"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func MirrorProposalsCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_MirrorProposals,
		"Starting mirror proposals callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	_, found := k.stakeibcKeeper.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(stakeibctypes.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	highestID, _ := k.GetProposalID(ctx, chainId)

	var proposals govtypesv1.Proposals
	codec.NewLegacyAmino().UnmarshalJSON(args, &proposals)

	// if err != nil {
	// 	return sdkerrors.Wrapf(stakeibctypes.ErrMarshalFailure, "unable to unmarshal query response into Proposal type, err: %s", err.Error())
	// }

	for _, proposal := range proposals {
		prop := *proposal
		if prop.Id > highestID {
			liquidProp := types.Proposal{HostZoneChainId: chainId, GovProposal: prop}
			k.SetProposal(ctx, liquidProp)
			k.SetProposalID(ctx, highestID+1, chainId)
			k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_MirrorProposals,
				"Added proposal #%d from host zone: %s VoteStartTime: %s VoteEndTime: %s", liquidProp.GovProposal.Id, liquidProp.HostZoneChainId, liquidProp.GovProposal.VotingStartTime, liquidProp.GovProposal.VotingEndTime))
		}
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_MirrorProposals, "Query response - Chain ID: %s Proposal ID: %d, VotingStartTime: %s, VotingEndTime: %s",
			chainId, prop.Id, prop.VotingStartTime, prop.VotingEndTime))
	}

	// query host zone unbonding period

	return nil
}
