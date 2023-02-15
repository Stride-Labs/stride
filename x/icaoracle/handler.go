package icaoracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// NewMessageHandler returns icaoracle module messages
func NewMessageHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		context := sdk.UnwrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgAddOracle:
			res, err := msgServer.AddOracle(context, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRestoreOracleICA:
			res, err := msgServer.RestoreOracleICA(context, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
