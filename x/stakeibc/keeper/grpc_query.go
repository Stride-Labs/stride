package keeper

import (
	"github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

var _ types.QueryServer = Keeper{}
