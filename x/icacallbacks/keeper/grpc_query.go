package keeper

import (
	"github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
