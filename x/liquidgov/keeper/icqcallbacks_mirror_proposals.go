package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/Stride-Labs/stride/v5/utils"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
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

	// get bytes of last entry in range query
	lastPropBytes := sdk.InclusiveEndBytes(args)
	lastQueriedProposal := govtypes.Proposal{}

	// Unmarshal the last query response bytes into a Proposal struct
	err := k.cdc.Unmarshal(lastPropBytes, &lastQueriedProposal)
	if err != nil {
		return sdkerrors.Wrapf(stakeibctypes.ErrMarshalFailure, "unable to unmarshal query response into Proposal type, err: %s", err.Error())
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_MirrorProposals, "Query response - Proposal Id: %s, VotingStartTime: %v, VotingEndTime: %v",
		lastQueriedProposal.Id, lastQueriedProposal.VotingStartTime, lastQueriedProposal.VotingEndTime))

	// // get most recent proposal id on Stride
	// newestId := k.GetNewestProposal(ctx, hostzone)
	// // if query response proposal ids > newest stride proposal
	// // add to stride proposal store
	// for _, proposal := range proposals {
	// 	if proposal.id > newestId {
	// 		// add stride specific proposal fields ie: hostzone

	// 		// add proposal to stride store
	// 		k.AddProposal(hostZone, proposal)
	// 	}
	// }
	return nil
}
