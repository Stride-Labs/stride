package types

import (
	"time"
)

var (
	DefaultClaimDenom                      = "ustrd"
	DefaultAirdropDuration                 = time.Hour * 24 * 30 * 12 * 3 // 3 years
	DefaultVestingDurationForDelegateStake = time.Hour * 24 * 30 * 2      // 2 months
	DefaultVestingDurationForLiquidStake   = time.Hour * 24 * 30 * 4      // 4 months
	DefaultAirdropIdentifier               = "stride"
)
