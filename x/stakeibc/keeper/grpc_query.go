package keeper

import (
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var _ types.QueryServer = Keeper{}
