package types

import (
	"time"
)

var (
	DefaultClaimDenom                      = "ustrd"
	DefaultAirdropDuration                 = time.Hour
	DefaultVestingDurationForDelegateStake = time.Hour * 2
	DefaultVestingDurationForLiquidStake   = time.Hour * 4
)
