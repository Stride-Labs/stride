package keeper

import (
	"github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

var _ types.QueryServer = Keeper{}
