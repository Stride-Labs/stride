package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	epochstypes "github.com/Stride-Labs/stride/v11/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v11/utils"
	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
)

const (
	GOV_STORE_QUERY_KEY            = "store/gov/key"
)

func (k Keeper) UpdateProposalICQ(ctx sdk.Context, hostZone stakeibctypes.HostZone, proposalID uint64) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for proposals to %s", hostZone.ChainId))

	// query for next proposal ID
	queryData := govtypes.ProposalKey(proposalID)

	// The query should timeout at the start of the next epoch
	ttl, err := k.stakeibcKeeper.GetStartTimeNextEpoch(ctx, epochstypes.STRIDE_EPOCH)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not get start time for next epoch: %s", err.Error())
	}

	if err := k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_UpdateProposals,
		hostZone.ChainId,
		hostZone.ConnectionId,
		// use "gov" store to access proposals which live in the gov module
		// use "key" suffix to retrieve proposals by key
		GOV_STORE_QUERY_KEY,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for proposals, error : %s", err.Error()))
		return err
	}
	return nil
}
