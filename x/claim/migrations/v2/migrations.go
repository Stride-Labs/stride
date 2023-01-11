package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldclaimtypes "github.com/Stride-Labs/stride/v4/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
)

func convertToNewClaimParams(oldParams oldclaimtypes.Params) claimtypes.Params {
	var newParams claimtypes.Params
	for _, airDrop := range oldParams.Airdrops {
		newAirDrop := claimtypes.Airdrop{
			AirdropIdentifier:  airDrop.AirdropIdentifier,
			AirdropStartTime:   airDrop.AirdropStartTime,
			AirdropDuration:    airDrop.AirdropDuration,
			ClaimDenom:         airDrop.ClaimDenom,
			DistributorAddress: airDrop.DistributorAddress,
			ClaimedSoFar:       sdk.NewInt(airDrop.ClaimedSoFar),
		}
		newParams.Airdrops = append(newParams.Airdrops, &newAirDrop)
	}
	return newParams
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	return nil
}
