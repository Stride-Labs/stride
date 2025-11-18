package keeper

import (
	"github.com/Stride-Labs/stride/v30/x/records/types"
)

var _ types.QueryServer = Keeper{}
