package keeper

import (
	"github.com/Stride-Labs/stride/v12/x/records/types"
)

var _ types.QueryServer = Keeper{}
