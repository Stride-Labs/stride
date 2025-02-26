package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v26/testutil/keeper"
	testkeeper "github.com/Stride-Labs/stride/v26/testutil/keeper"
	"github.com/Stride-Labs/stride/v26/testutil/nullify"
	epochtypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestHostZoneQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNHostZone(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetHostZoneRequest
		response *types.QueryGetHostZoneResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetHostZoneRequest{ChainId: msgs[0].ChainId},
			response: &types.QueryGetHostZoneResponse{HostZone: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetHostZoneRequest{ChainId: msgs[1].ChainId},
			response: &types.QueryGetHostZoneResponse{HostZone: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetHostZoneRequest{ChainId: strconv.Itoa((len(msgs)))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.HostZone(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestHostZoneQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNHostZone(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllHostZoneRequest {
		return &types.QueryAllHostZoneRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.HostZoneAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.HostZone), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.HostZone),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.HostZoneAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.HostZone), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.HostZone),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.HostZoneAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.HostZone),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.HostZoneAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func TestValidatorQuery(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	validatorsByHostZone := make(map[string][]*types.Validator)
	validators := []*types.Validator{}
	nullify.Fill(&validators)

	chainId := "GAIA"
	hostZone := &types.HostZone{
		ChainId:    chainId,
		Validators: validators,
	}
	nullify.Fill(&hostZone)
	validatorsByHostZone[chainId] = validators
	keeper.SetHostZone(ctx, *hostZone)

	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetValidatorsRequest
		response *types.QueryGetValidatorsResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetValidatorsRequest{ChainId: chainId},
			response: &types.QueryGetValidatorsResponse{Validators: validators},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Validators(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestAddressUnbondings(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	// Setup DayEpoch Tracker for current epoch 100
	const nanosecondsInDay = 86400000000000
	const testTimeNanos = 1704067200000000000 // 2024-01-01 00:00:00 is start of epoch 100
	dayEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        100,
		NextEpochStartTime: testTimeNanos + nanosecondsInDay,
		Duration:           nanosecondsInDay,
	}
	keeper.SetEpochTracker(ctx, dayEpochTracker)

	// Setup HostZones with different unbonding periods
	cosmosZone := types.HostZone{
		ChainId:         "cosmos",
		UnbondingPeriod: 21,
	}
	keeper.SetHostZone(ctx, cosmosZone)
	strideZone := types.HostZone{
		ChainId:         "stride",
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
			Id:                "stride.101.strideAddrUserA",
			Receiver:          "strideAddrUserA",
			NativeTokenAmount: sdk.NewInt(2000),
			Denom:             "ustrd",
			HostZoneId:        "stride",
			EpochNumber:       uint64(101),
			ClaimIsPending:    false,
		},
		{
			Id:                "stride.110.strideAddrUserA",
			Receiver:          "strideAddrUserA",
			NativeTokenAmount: sdk.NewInt(5000),
			Denom:             "ustrd",
			HostZoneId:        "stride",
			EpochNumber:       uint64(110),
			ClaimIsPending:    false,
		},
		{
			Id:                "stride.101.strideAddrUserB",
			Receiver:          "strideAddrUserB",
			NativeTokenAmount: sdk.NewInt(8500),
			Denom:             "ustrd",
			HostZoneId:        "stride",
			EpochNumber:       uint64(101),
			ClaimIsPending:    false,
		},
		{
			Id:                "cosmos.101.cosmosAddrUserA",
			Receiver:          "cosmosAddrUserA",
			NativeTokenAmount: sdk.NewInt(1200),
			Denom:             "uatom",
			HostZoneId:        "cosmos",
			EpochNumber:       uint64(101),
			ClaimIsPending:    false,
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
				AddressUnbondings: []types.AddressUnbonding{},
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
						Address:                "cosmosAddrUserA",
						Receiver:               "cosmosAddrUserA",
						Amount:                 sdk.NewInt(1200),
						Denom:                  "uatom",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
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
						Address:                "strideAddrUserA",
						Receiver:               "strideAddrUserA",
						Amount:                 sdk.NewInt(2000),
						Denom:                  "ustrd",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
					{
						Address:                "strideAddrUserA",
						Receiver:               "strideAddrUserA",
						Amount:                 sdk.NewInt(5000),
						Denom:                  "ustrd",
						EpochNumber:            uint64(110),
						ClaimIsPending:         false,
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
						Address:                "strideAddrUserB",
						Receiver:               "strideAddrUserB",
						Amount:                 sdk.NewInt(8500),
						Denom:                  "ustrd",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
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
						Address:                "strideAddrUserA",
						Receiver:               "strideAddrUserA",
						Amount:                 sdk.NewInt(2000),
						Denom:                  "ustrd",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
					{
						Address:                "cosmosAddrUserA",
						Receiver:               "cosmosAddrUserA",
						Amount:                 sdk.NewInt(1200),
						Denom:                  "uatom",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
						UnbondingEstimatedTime: "2024-01-27 00:00:00 +0000 UTC",
					},
					{
						Address:                "strideAddrUserA",
						Receiver:               "strideAddrUserA",
						Amount:                 sdk.NewInt(5000),
						Denom:                  "ustrd",
						EpochNumber:            uint64(110),
						ClaimIsPending:         false,
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
			err:      status.Error(codes.InvalidArgument, "invalid request"),
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

func TestEpochTrackerQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNEpochTracker(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetEpochTrackerRequest
		response *types.QueryGetEpochTrackerResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: msgs[0].EpochIdentifier,
			},
			response: &types.QueryGetEpochTrackerResponse{EpochTracker: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: msgs[1].EpochIdentifier,
			},
			response: &types.QueryGetEpochTrackerResponse{EpochTracker: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.EpochTracker(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestAllEpochTrackerQuery(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNEpochTracker(keeper, ctx, 5)

	resp, err := keeper.EpochTrackerAll(wctx, &types.QueryAllEpochTrackerRequest{})
	require.NoError(t, err)
	require.Len(t, resp.EpochTracker, 5)
	require.Subset(t,
		nullify.Fill(msgs),
		nullify.Fill(resp.EpochTracker),
	)
}

func (s *KeeperTestSuite) TestNextPacketSequenceQuery() {
	portId := "transfer"
	channelId := "channel-0"
	sequence := uint64(10)
	context := sdk.WrapSDKContext(s.Ctx)

	// Set a channel sequence
	s.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(s.Ctx, portId, channelId, sequence)

	// Test a successful query
	response, err := s.App.StakeibcKeeper.NextPacketSequence(context, &types.QueryGetNextPacketSequenceRequest{
		ChannelId: channelId,
		PortId:    portId,
	})
	s.Require().NoError(err)
	s.Require().Equal(sequence, response.Sequence)

	// Test querying a non-existent channel (should fail)
	_, err = s.App.StakeibcKeeper.NextPacketSequence(context, &types.QueryGetNextPacketSequenceRequest{
		ChannelId: "fake-channel",
		PortId:    portId,
	})
	s.Require().ErrorContains(err, "channel and port combination not found")
}
