package epochs

import (
	"fmt"

	"github.com/Stride-Labs/stride/v3/x/epochs/keeper"
	"github.com/Stride-Labs/stride/v3/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for epochs module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, fmt.Errorf("Unknown request: %s", errMsg)
	}
}
