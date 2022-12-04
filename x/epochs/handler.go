package epochs

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/epochs/keeper"
	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

// NewHandler returns a handler for epochs module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, fmt.Errorf("%s: unknown request", errMsg)
	}
}
