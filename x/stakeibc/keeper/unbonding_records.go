package keeper

import (
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetValidator set validator in the store
func (k Keeper) CreateEpochUndelegations(ctx sdk.Context, epochNumber int64) bool {
	hostZoneUnbondings := []*recordstypes.EpochUnbondingRecordHostZoneUnbonding{}
	addEpochUndelegation := func(index int64, hostZone types.HostZone) (stop bool) {
		hostZoneUnbonding := recordstypes.EpochUnbondingRecordHostZoneUnbonding{
			Amount:     uint64(0),
			Denom:      hostZone.HostDenom,
			HostZoneId: hostZone.ChainId,
		}
		hostZoneUnbondings = append(hostZoneUnbondings, &hostZoneUnbonding)
		return false
	}
	k.IterateHostZones(ctx, addEpochUndelegation)
	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: hostZoneUnbondings,
	}
	k.recordsKeeper.AppendEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return true
}
