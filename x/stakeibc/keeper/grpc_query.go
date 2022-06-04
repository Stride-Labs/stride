package keeper

import (
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

var _ types.QueryServer = Keeper{}
