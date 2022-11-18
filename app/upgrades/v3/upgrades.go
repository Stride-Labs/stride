package v3

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimkeeper "github.com/Stride-Labs/stride/v3/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v3/x/claim/types"
	recordskeeper "github.com/Stride-Labs/stride/v3/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v3/x/records/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName         = "v3"
	airdropDistributors = []string{
		"stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l",
		"stride104az7rd5yh3p8qn4ary8n3xcwquuwgee4vnvvc",
		"stride17kvetgthwt6caku5qjs2rx2njgh26vmg448r5u",
		"stride1swvv9kpp75e60pvlv5x6mcw5f54qgpph239e5s",
		"stride1ywrhas3ae7z3ljqxmgdzjx8wyaf3djwuh4hdlj",
	}
	airdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
	airdropDuration    = time.Hour * 24 * 30 * 12 * 3 // 3 years
)

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ck claimkeeper.Keeper,
	rk recordskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		newVm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVm, err
		}

		// total number of airdrop distributors must be equal to identifiers
		if len(airdropDistributors) == len(airdropIdentifiers) {
			for idx, airdropDistributor := range airdropDistributors {
				err = ck.CreateAirdropAndEpoch(ctx, airdropDistributor, claimtypes.DefaultClaimDenom, uint64(ctx.BlockTime().Unix()), uint64(airdropDuration.Seconds()), airdropIdentifiers[idx])
				if err != nil {
					return newVm, err
				}
			}
		}
		ck.LoadAllocationData(ctx, allocations)

		// revert juno deposit records (currently stuck due to timeout)
		badRecords := []uint64{848, 851, 860, 863, 902, 908, 914, 920, 927, 934, 940, 946, 952, 958, 965, 972, 979,
			986, 992, 998, 1005, 1012, 1019, 1026, 1033, 1040, 1047, 1054, 1061, 1068, 1074, 1080,
			1086, 1092, 1098, 1104, 1110, 1116, 1112, 1128, 1140, 1146, 1152, 1158, 1165, 1171, 1189, 1195, 1201, 1207, 1213, 1217, 1220, 1226, 1232, 1238, 1244}
		for _, recordId := range badRecords {
			depositRecord, found := rk.GetDepositRecord(ctx, recordId)
			if found {
				if (depositRecord.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS) && (depositRecord.HostZoneId == "juno-1") {
					depositRecord.Status = recordstypes.DepositRecord_DELEGATION_QUEUE
					rk.SetDepositRecord(ctx, depositRecord)
				}
			}
		}

		return newVm, nil
	}
}
