package keeper

import (
	"github.com/Stride-Labs/stride/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
