package v2

import (
	sdkmath "cosmossdk.io/math"

	oldclaimtypes "github.com/Stride-Labs/stride/v9/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
)

func convertToNewAirdrop(oldAirdrop oldclaimtypes.Airdrop) claimtypes.Airdrop {
	return claimtypes.Airdrop{
		AirdropIdentifier:  oldAirdrop.AirdropIdentifier,
		AirdropStartTime:   oldAirdrop.AirdropStartTime,
		AirdropDuration:    oldAirdrop.AirdropDuration,
		ClaimDenom:         oldAirdrop.ClaimDenom,
		DistributorAddress: oldAirdrop.DistributorAddress,
		ClaimedSoFar:       sdkmath.NewInt(oldAirdrop.ClaimedSoFar),
	}
}

func convertToNewClaimParams(oldParams oldclaimtypes.Params) claimtypes.Params {
	var newParams claimtypes.Params
	for _, oldAirdrop := range oldParams.Airdrops {
		newAirDrop := convertToNewAirdrop(*oldAirdrop)
		newParams.Airdrops = append(newParams.Airdrops, &newAirDrop)
	}
	return newParams
}
