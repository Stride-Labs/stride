package keeper

import (
	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

var _ types.QueryServer = Keeper{}
