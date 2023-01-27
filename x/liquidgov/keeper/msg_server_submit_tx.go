package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

func (k Keeper) MirrorProposals(ctx sdk.Context, hostZone stakeibctypes.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for proposals to %s", hostZone.ChainId))

	// Ensure ICQ can be issued now! else fail the callback
	valid, err := k.stakeibcKeeper.IsWithinBufferWindow(ctx)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
	} else if !valid {
		return sdkerrors.Wrapf(stakeibctypes.ErrOutsideIcqWindow, "outside the buffer time during which ICQs are allowed (%s)", hostZone.HostDenom)
	}

	endTime := ctx.BlockTime().AddDate(1, 0, 0) // TODO veryify range query works, then change logic to compare to latest proposal on stride

	queryData := govtypes.ActiveProposalByTimeKey(endTime)

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
		icqtypes.GOV_STORE_QUERY_WITH_PROOF,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return err
	}
	return nil
}
