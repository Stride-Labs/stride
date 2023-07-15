package keeper

import (
	"github.com/Stride-Labs/stride/v11/x/records/types"
)

var _ types.QueryServer = Keeper{}
