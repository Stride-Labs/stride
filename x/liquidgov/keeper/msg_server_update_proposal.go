package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateProposal(goCtx context.Context, msg *types.MsgUpdateProposal) (*types.MsgUpdateProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Keeper.Logger(ctx).Info("Updating Proposal %d -- about to kick off ICQ", msg.ProposalId)

	hostZone, _ := k.stakeibcKeeper.GetHostZone(ctx, msg.HostZoneId)

	k.Keeper.UpdateProposalICQ(ctx, hostZone, msg.ProposalId)

	return &types.MsgUpdateProposalResponse{}, nil
}
