package interchainquery

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v26/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/v26/x/interchainquery/types"
)

// NewHandler returns a handler for interchainquery module messages
func NewHandler(k keeper.Keeper) baseapp.MsgServiceHandler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}
}
