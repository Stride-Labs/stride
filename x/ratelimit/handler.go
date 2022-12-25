package ratelimit

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// NewMessageHandler returns ratelimit module messages
func NewMessageHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		_ = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// NewRateLimitProposalHandler returns ratelimit module's proposals
func NewRateLimitProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddRateLimitProposal:
			return handleAddRateLimitProposal(ctx, k, c)
		case *types.UpdateRateLimitProposal:
			return handleUpdateRateLimitProposal(ctx, k, c)
		case *types.RemoveRateLimitProposal:
			return handleRemoveRateLimitProposal(ctx, k, c)
		case *types.ResetRateLimitProposal:
			return handleResetRateLimitProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ratelimit proposal content type: %T", c)
		}
	}
}

// Handler for adding a rate limit through governance
func handleAddRateLimitProposal(ctx sdk.Context, k keeper.Keeper, proposal *types.AddRateLimitProposal) error {
	return k.GovAddRateLimit(ctx, proposal)
}

// Handler for updating a rate limit through governance
func handleUpdateRateLimitProposal(ctx sdk.Context, k keeper.Keeper, proposal *types.UpdateRateLimitProposal) error {
	return k.GovUpdateRateLimit(ctx, proposal)
}

// Handler for removing a rate limit through governance
func handleRemoveRateLimitProposal(ctx sdk.Context, k keeper.Keeper, proposal *types.RemoveRateLimitProposal) error {
	return k.GovRemoveRateLimit(ctx, proposal)
}

// Handler for resetting a rate limit through governance
func handleResetRateLimitProposal(ctx sdk.Context, k keeper.Keeper, proposal *types.ResetRateLimitProposal) error {
	return k.GovResetRateLimit(ctx, proposal)
}
