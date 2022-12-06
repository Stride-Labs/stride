package keeper

import (
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/v3/x/ibcratelimit/types"
)

func CheckRateLimit(direction types.PacketDirection, packet exported.PacketI,
) error {
	// TODO
	return nil
}
