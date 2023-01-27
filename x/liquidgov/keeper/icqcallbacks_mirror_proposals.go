package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
)

func MirrorProposalsCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	// get most recent proposal id on Stride
	newestId := k.GetNewestProposal(ctx, hostzone)
	// if query response proposal ids > newest stride proposal
	// add to stride proposal store
	for _, proposal := range proposals {
		if proposal.id > newestId {
			// add stride specific proposal fields ie: hostzone

			// add proposal to stride store
			k.AddProposal(hostZone, proposal)
		}
	}
}
