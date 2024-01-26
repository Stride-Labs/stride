package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v18/testutil/keeper"
	epochtypes "github.com/Stride-Labs/stride/v18/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v18/x/records/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

func TestAddressUnbondings(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	// Setup DayEpoch Tracker for current epoch 100
	const nanosecondsInDay = 86400000000000	
	const testTimeNanos = 1704067200000000000 // 2024-01-01 00:00:00 is start of epoch 100
	dayEpochTracker := types.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber: 100,
		NextEpochStartTime: testTimeNanos + nanosecondsInDay,
		Duration: nanosecondsInDay,
	}
	keeper.SetEpochTracker(ctx, dayEpochTracker)

	// Setup HostZones with different unbonding periods
	cosmosZone := types.HostZone{
		ChainId: "cosmos",
		UnbondingPeriod: 21,
	}
	keeper.SetHostZone(ctx, cosmosZone)
	strideZone := types.HostZone{
		ChainId: "stride",
		UnbondingPeriod: 9, // just so different unbonding period
	}
	keeper.SetHostZone(ctx, strideZone)

	// Setup some epoch unbonding records to test against
	for _, epochUnbondingRecord := range []*recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 101,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId: "stride",
					UserRedemptionRecords: []string{
						"stride.101.strideAddrUserA",
						"stride.101.strideAddrUserB",
					},
				},
				{
					HostZoneId: "cosmos",
					UserRedemptionRecords: []string{
						"cosmos.101.cosmosAddrUserA",
					},
				},			
			},
		},
		{
			EpochNumber: 110,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId: "stride",
					UserRedemptionRecords: []string{
						"stride.110.strideAddrUserA",
					},
				},			
			},
		},		
	} {
		keeper.RecordsKeeper.SetEpochUnbondingRecord(ctx, *epochUnbondingRecord)
	}

	// Setup corresponding user unbonding records to test with
	for _, userRedemptionRecord := range []*recordtypes.UserRedemptionRecord{
		{
			Id: "stride.101.strideAddrUserA",
			Receiver: "strideAddrUserA",
			NativeTokenAmount: sdk.NewInt(2000),
			Denom: "ustrd",
			HostZoneId: "stride",
			EpochNumber: uint64(101),
			ClaimIsPending: false,
		},
		{
			Id: "stride.110.strideAddrUserA",
			Receiver: "strideAddrUserA",
			NativeTokenAmount: sdk.NewInt(5000),
			Denom: "ustrd",
			HostZoneId: "stride",
			EpochNumber: uint64(110),
			ClaimIsPending: false,
		},
		{
			Id: "stride.101.strideAddrUserB",
			Receiver: "strideAddrUserB",
			NativeTokenAmount: sdk.NewInt(8500),
			Denom: "ustrd",
			HostZoneId: "stride",
			EpochNumber: uint64(101),
			ClaimIsPending: false,
		},
		{
			Id: "cosmos.101.cosmosAddrUserA",
			Receiver: "cosmosAddrUserA",
			NativeTokenAmount: sdk.NewInt(1200),
			Denom: "uatom",
			HostZoneId: "cosmos",
			EpochNumber: uint64(101),
			ClaimIsPending: false,
		},
	} {
		keeper.RecordsKeeper.SetUserRedemptionRecord(ctx, *userRedemptionRecord)
	}

	// Test cases:
	// Single Address --> no redemption records found
	// Single Address --> single redemption record found
	// Single Address --> multiple redemption records across different epochs
	// Multiple Addresses --> find record for only one address but not the others
	// Multiple Addresses --> find user records for each, more than one for some
	// Invalid query or address --> expected err case

	for _, tc := range []struct {
		desc     string
		request  *types.QueryAddressUnbondings
		response *types.QueryAddressUnbondingsResponse
		err      error
	}{
		{
			desc: "Single input address without any records expected",
			request: &types.QueryAddressUnbondings{
				Address: "cosmosAddrUserB",
			},
			response: &types.QueryAddressUnbondingsResponse{
				AddressUnbondings: []types.AddressUnbonding{
				},
			},
		},
		{
			desc: "Single input address with one record expected",
			request: &types.QueryAddressUnbondings{
				Address: "cosmosAddrUserA",
			},
			response: &types.QueryAddressUnbondingsResponse{
				AddressUnbondings: []types.AddressUnbonding{
					{
						Address: "cosmosAddrUserA",
						Receiver: "cosmosAddrUserA",
						Amount: sdk.NewInt(1200),
						Denom: "uatom",
						EpochNumber: uint64(101),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-27 00:00:00 +0000 UTC",
					},	
				},
			},
		},		
		{
			desc: "Single input address with multiple records across epochs",
			request: &types.QueryAddressUnbondings{
				Address: "strideAddrUserA",
			},
			response: &types.QueryAddressUnbondingsResponse{
				AddressUnbondings: []types.AddressUnbonding{
					{
						Address: "strideAddrUserA",
						Receiver: "strideAddrUserA",
						Amount: sdk.NewInt(2000),
						Denom: "ustrd",
						EpochNumber: uint64(101),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
					{
						Address: "strideAddrUserA",
						Receiver: "strideAddrUserA",
						Amount: sdk.NewInt(5000),
						Denom: "ustrd",
						EpochNumber: uint64(110),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},					
				},
			},
		},
		{
			desc: "Multiple input addresses only one has a record, others unfound",
			request: &types.QueryAddressUnbondings{
				Address: "cosmosAddrUserB,strideAddrUserB,strideAddrUserC",
			},
			response: &types.QueryAddressUnbondingsResponse{
				AddressUnbondings: []types.AddressUnbonding{
					{
						Address: "strideAddrUserB",
						Receiver: "strideAddrUserB",
						Amount: sdk.NewInt(8500),
						Denom: "ustrd",
						EpochNumber: uint64(101),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
				},
			},
		},
		{
			desc: "Multiple input addresses all with one or more records",
			request: &types.QueryAddressUnbondings{
				Address: "strideAddrUserA, cosmosAddrUserA",
			},
			response: &types.QueryAddressUnbondingsResponse{
				AddressUnbondings: []types.AddressUnbonding{
					{
						Address: "strideAddrUserA",
						Receiver: "strideAddrUserA",
						Amount: sdk.NewInt(2000),
						Denom: "ustrd",
						EpochNumber: uint64(101),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
					{
						Address: "cosmosAddrUserA",
						Receiver: "cosmosAddrUserA",
						Amount: sdk.NewInt(1200),
						Denom: "uatom",
						EpochNumber: uint64(101),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-27 00:00:00 +0000 UTC",
					},					
					{
						Address: "strideAddrUserA",
						Receiver: "strideAddrUserA",
						Amount: sdk.NewInt(5000),
						Denom: "ustrd",
						EpochNumber: uint64(110),
						ClaimIsPending: false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},						
				},
			},
		},
		{
			desc: "No address given, error expected",
			request: &types.QueryAddressUnbondings{
				Address: "",
			},
			response: nil,
			err: status.Error(codes.InvalidArgument, "invalid request"),
		},								
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.AddressUnbondings(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					tc.response,
					response,
				)
			}
		})
	}
}
