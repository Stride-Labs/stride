package keeper

import (
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func CheckRateLimit(direction types.PacketDirection, packet exported.PacketI,
) error {
	// TODO
	return nil
}
