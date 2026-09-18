package v34

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	recordskeeper "github.com/Stride-Labs/stride/v34/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

const CosmosHubChainId = "cosmoshub-4"

// StrandedLsmDeposit identifies one LSM token deposit and the amount/validator it is expected to
// still carry, so CloseCosmosHubLsmDeposit can refuse to act if chain state has drifted from the
// constant.
type StrandedLsmDeposit struct {
	Denom            string
	ValidatorAddress string
	Amount           sdkmath.Int
}

// CosmosHubStrandedLsmDeposit is the one Cosmos Hub LSM token deposit whose detokenize executed
// on the Hub but whose success acknowledgement was stranded on a closed channel, leaving
// Stride's record stuck in DETOKENIZATION_FAILED forever even though the underlying stake is
// real and already delegated.
//
// Measured 2026-09-17/18: the Hub delegation ICA's on-chain delegation to stakewithus
// (cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr) exceeds the tracked validator
// delegation by exactly 10,999,999 uatom. The LSM deposit with this denom sits in
// DETOKENIZATION_FAILED (deposit_id ba82f49a967dda66331221b18018d305198abd72fba55552a451f1012274a57d,
// staker stride1v6ll7lj9qeyf6g6at2c4vqetxg5877558cqpjj): the detokenize executed on the Hub but
// its ack was stranded on a closed channel and the retry failed.
//
// !!! RE-VERIFY right before the proposal !!!
var CosmosHubStrandedLsmDeposit = StrandedLsmDeposit{
	Denom:            "cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr/114571",
	ValidatorAddress: "cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr", // stakewithus
	Amount:           sdkmath.NewInt(10_999_999),                             // uatom
}

// CloseCosmosHubLsmDeposit replays the detokenize success path (see
// x/stakeibc/keeper/icacallbacks_detokenize.go's DetokenizeCallback) for
// CosmosHubStrandedLsmDeposit: the stake already exists on the Hub, so this only needs to move
// Stride's bookkeeping from "pending LSM deposit" to "native delegation."
//
// Order matters and nothing is written until every check passes:
//  1. Host zone cosmoshub-4 present, else log + return false (non-mainnet).
//  2. The LSM deposit is found by (chain id, denom) and its status, amount and validator address
//     match the constant exactly.
//  3. The validator is present on the host zone.
//
// Once every check passes: remove the LSM deposit, add its amount to the validator's (and the
// host zone's) delegation exactly as a successful detokenize ack would, and persist the host
// zone. Any mismatch is logged as an error and the function returns false with nothing written;
// it never returns an error, since halting the chain is disproportionate for an accounting fix.
func CloseCosmosHubLsmDeposit(ctx sdk.Context, sk stakeibckeeper.Keeper, rk recordskeeper.Keeper) (applied bool) {
	hostZone, found := sk.GetHostZone(ctx, CosmosHubChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v34: host zone %s not found, skipping LSM deposit close-out", CosmosHubChainId))
		return false
	}

	deposit, found := rk.GetLSMTokenDeposit(ctx, CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("v34: %s LSM deposit %s not found; close-out NOT applied, "+
			"re-verify constants and close out in a later upgrade", CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom))
		return false
	}
	if deposit.Status != recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED {
		ctx.Logger().Error(fmt.Sprintf("v34: %s LSM deposit %s has status %s, expected %s; close-out NOT applied, "+
			"re-verify constants and close out in a later upgrade", CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom,
			deposit.Status, recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED))
		return false
	}
	if !deposit.Amount.Equal(CosmosHubStrandedLsmDeposit.Amount) {
		ctx.Logger().Error(fmt.Sprintf("v34: %s LSM deposit %s has amount %v, expected %v; close-out NOT applied, "+
			"re-verify constants and close out in a later upgrade", CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom,
			deposit.Amount, CosmosHubStrandedLsmDeposit.Amount))
		return false
	}
	if deposit.ValidatorAddress != CosmosHubStrandedLsmDeposit.ValidatorAddress {
		ctx.Logger().Error(fmt.Sprintf("v34: %s LSM deposit %s has validator %s, expected %s; close-out NOT applied, "+
			"re-verify constants and close out in a later upgrade", CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom,
			deposit.ValidatorAddress, CosmosHubStrandedLsmDeposit.ValidatorAddress))
		return false
	}
	if _, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, CosmosHubStrandedLsmDeposit.ValidatorAddress); !found {
		ctx.Logger().Error(fmt.Sprintf("v34: validator %s not found on %s; close-out NOT applied, "+
			"re-verify constants and close out in a later upgrade", CosmosHubStrandedLsmDeposit.ValidatorAddress, CosmosHubChainId))
		return false
	}

	// Every check passed: replay the detokenize success path exactly (see DetokenizeCallback)
	rk.RemoveLSMTokenDeposit(ctx, CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom)
	if err := sk.AddDelegationToValidator(ctx, &hostZone, CosmosHubStrandedLsmDeposit.ValidatorAddress,
		CosmosHubStrandedLsmDeposit.Amount, stakeibckeeper.ICACallbackID_Detokenize); err != nil {
		// Unreachable given the checks above (the validator is present and the amount is
		// positive), but AddDelegationToValidator can error, so handle it without a panic or an
		// upgrade error.
		ctx.Logger().Error(fmt.Sprintf("v34: unable to add delegation while closing %s LSM deposit %s: %s; "+
			"close-out NOT fully applied", CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom, err))
		return false
	}
	sk.SetHostZone(ctx, hostZone)

	ctx.Logger().Info(fmt.Sprintf("v34: %s LSM deposit %s closed: %v %s booked to validator %s",
		CosmosHubChainId, CosmosHubStrandedLsmDeposit.Denom, CosmosHubStrandedLsmDeposit.Amount,
		hostZone.HostDenom, CosmosHubStrandedLsmDeposit.ValidatorAddress))
	return true
}
