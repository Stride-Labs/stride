package v35

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	recordskeeper "github.com/Stride-Labs/stride/v35/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v35/x/records/types"
)

// The one LSM deposit stuck in DETOKENIZATION_FAILED on mainnet (spec §3). Its detokenization
// (Hub tx 2161106F4332619D776FF664EA3F31BEDD21F93C1948C560791C51ACFA14DCA3, 2026-06-05) failed
// with "67850951.997832703528686576: not enough delegation shares": the tokenize-share record
// holds a fraction of a share less than the tokens minted, the same rounding failure v23, v25
// and v32 fixed by retrying with one token less
const (
	FailedLSMDepositChainId = "cosmoshub-4"
	FailedLSMDepositDenom   = "cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327"
)

var FailedLSMDepositAmount = sdkmath.NewInt(67_850_952)

// ResetFailedLSMDeposit requeues the failed deposit for detokenization with its amount reduced
// by one, so the EndBlocker retries it and the redeem fits within the record's shares. The retry
// books the tokens as a delegation (already counted in the redemption rate), which the drain then
// unbonds. It only acts on the exact record measured, so it cannot touch anything else or apply
// twice; anything unexpected is logged and skipped, since the amount is immaterial
func ResetFailedLSMDeposit(ctx sdk.Context, k recordskeeper.Keeper) {
	deposit, found := k.GetLSMTokenDeposit(ctx, FailedLSMDepositChainId, FailedLSMDepositDenom)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: LSM deposit %s not found, skipping reset", FailedLSMDepositDenom))
		return
	}
	if deposit.Status != recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED || !deposit.Amount.Equal(FailedLSMDepositAmount) {
		ctx.Logger().Error(fmt.Sprintf("v35: LSM deposit %s is %s with amount %v, expected %s with amount %v; skipping reset",
			FailedLSMDepositDenom, deposit.Status, deposit.Amount,
			recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, FailedLSMDepositAmount))
		return
	}

	deposit.Status = recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE
	deposit.Amount = deposit.Amount.Sub(sdkmath.OneInt())
	k.SetLSMTokenDeposit(ctx, deposit)
	ctx.Logger().Info(fmt.Sprintf("v35: LSM deposit %s requeued for detokenization with amount %v", FailedLSMDepositDenom, deposit.Amount))
}
