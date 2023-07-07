package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultClaimDenom                      = "ustrd"
	DefaultEpochDuration                   = time.Hour * 24 * 30                       // 1 month
	DefaultAirdropDuration                 = time.Hour * 24 * 30 * 12 * 3              // 3 years
	DefaultAirdropDailyLimit               = sdk.NewIntFromUint64(uint64(100_000_000)) // 100M
	DefaultVestingDurationForDelegateStake = time.Hour * 24 * 30 * 3                   // 3 months
	DefaultVestingDurationForLiquidStake   = time.Hour * 24 * 30 * 3                   // 3 months
	DefaultVestingInitialPeriod            = time.Hour * 24 * 30 * 3                   // 3 months
	DefaultAirdropIdentifier               = "stride"
)
