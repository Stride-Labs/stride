package autopilot

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v10/x/autopilot/keeper"
	"github.com/Stride-Labs/stride/v10/x/autopilot/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns autopilot module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}
}
