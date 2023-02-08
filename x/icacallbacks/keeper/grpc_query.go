package keeper

import (
	"github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
