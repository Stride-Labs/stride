package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) TransferLSMTokenToHost(ctx sdk.Context, lsmTokenDeposit types.LSMTokenDeposit) error {
	// TODO [LSM]
	return nil
}
