package keeper

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"

	"github.com/cosmos/cosmos-sdk/codec"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

func (k Keeper) MirrorProposals(ctx sdk.Context, hostZone stakeibctypes.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for proposals to %s", hostZone.ChainId))

	// query custom proposals path
	path := strings.Join([]string{"custom", govtypes.QuerierRoute, v1beta1.QueryProposals}, "/")
	// query data 1 page, default limit, voting period, no owner, no depositor
	queryData := codec.NewLegacyAmino().MustMarshalJSON(v1.NewQueryProposalsParams(1, 0, v1.StatusVotingPeriod, nil, nil))

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
		path,
		queryData,
		ttl,
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return err
	}
	return nil
}
