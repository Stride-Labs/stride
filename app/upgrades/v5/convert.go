package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	claimv1types "github.com/Stride-Labs/stride/v4/x/claim/types/v1"
)

func convertToNewClaimParams(oldProp claimv1types.Params) claimtypes.Params {
	var newParams claimtypes.Params
	for _, airDrop := range(oldProp.Airdrops) {
		newAirDrop := claimtypes.Airdrop{
			AirdropIdentifier: airDrop.AirdropIdentifier,
			AirdropStartTime: airDrop.AirdropStartTime,
			AirdropDuration: airDrop.AirdropDuration,
			ClaimDenom: airDrop.ClaimDenom,
			DistributorAddress: airDrop.DistributorAddress,
			ClaimedSoFar: sdk.NewInt(airDrop.ClaimedSoFar),
		}
		newParams.Airdrops = append(newParams.Airdrops, &newAirDrop)
	}
	return newParams
}