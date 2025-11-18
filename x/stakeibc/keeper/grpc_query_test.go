package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	epochtypes "github.com/Stride-Labs/stride/v30/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v30/x/records/types"
	"github.com/Stride-Labs/stride/v30/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestParamsQuery() {
	params := types.DefaultParams()
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	response, err := s.App.StakeibcKeeper.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}

func (s *KeeperTestSuite) TestHostZoneQuerySingle() {
	msgs := s.createNHostZone(2)
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
		s.Run(tc.desc, func() {
			response, err := s.App.StakeibcKeeper.HostZone(s.Ctx, tc.request)
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(
					tc.response,
					response,
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestHostZoneQueryPaginated() {
	msgs := s.createNHostZone(5)

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
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.StakeibcKeeper.HostZoneAll(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.HostZone), step)
			s.Require().Subset(
				msgs,
				resp.HostZone,
			)
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.StakeibcKeeper.HostZoneAll(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.HostZone), step)
			s.Require().Subset(
				msgs,
				resp.HostZone,
			)
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.App.StakeibcKeeper.HostZoneAll(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(msgs), int(resp.Pagination.Total))
		s.Require().ElementsMatch(
			msgs,
			resp.HostZone,
		)
	})
	s.Run("InvalidRequest", func() {
		_, err := s.App.StakeibcKeeper.HostZoneAll(s.Ctx, nil)
		s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func (s *KeeperTestSuite) TestValidatorQuery() {
	validatorsByHostZone := make(map[string][]*types.Validator)
	validators := []*types.Validator{}

	chainId := "GAIA"
	hostZone := &types.HostZone{
		ChainId:    chainId,
		Validators: validators,
	}
	validatorsByHostZone[chainId] = validators
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, *hostZone)

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
		s.Run(tc.desc, func() {
			response, err := s.App.StakeibcKeeper.Validators(s.Ctx, tc.request)
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(
					tc.response,
					response,
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAddressUnbondings() {
	// Setup DayEpoch Tracker for current epoch 100
	const nanosecondsInDay = 86400000000000
	const testTimeNanos = 1704067200000000000 // 2024-01-01 00:00:00 is start of epoch 100
	dayEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        100,
		NextEpochStartTime: testTimeNanos + nanosecondsInDay,
		Duration:           nanosecondsInDay,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, dayEpochTracker)

	// Setup HostZones with different unbonding periods
	cosmosZone := types.HostZone{
		ChainId:         "cosmos",
		UnbondingPeriod: 21,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, cosmosZone)
	strideZone := types.HostZone{
		ChainId:         "stride",
		UnbondingPeriod: 9, // just so different unbonding period
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, strideZone)

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
		s.App.StakeibcKeeper.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, *epochUnbondingRecord)
	}

	// Setup corresponding user unbonding records to test with
	for _, userRedemptionRecord := range []*recordtypes.UserRedemptionRecord{
		{
			Id:                "stride.101.strideAddrUserA",
			Receiver:          "strideAddrUserA",
			NativeTokenAmount: sdkmath.NewInt(2000),
			Denom:             "ustrd",
			HostZoneId:        "stride",
			EpochNumber:       uint64(101),
			ClaimIsPending:    false,
		},
		{
			Id:                "stride.110.strideAddrUserA",
			Receiver:          "strideAddrUserA",
			NativeTokenAmount: sdkmath.NewInt(5000),
			Denom:             "ustrd",
			HostZoneId:        "stride",
			EpochNumber:       uint64(110),
			ClaimIsPending:    false,
		},
		{
			Id:                "stride.101.strideAddrUserB",
			Receiver:          "strideAddrUserB",
			NativeTokenAmount: sdkmath.NewInt(8500),
			Denom:             "ustrd",
			HostZoneId:        "stride",
			EpochNumber:       uint64(101),
			ClaimIsPending:    false,
		},
		{
			Id:                "cosmos.101.cosmosAddrUserA",
			Receiver:          "cosmosAddrUserA",
			NativeTokenAmount: sdkmath.NewInt(1200),
			Denom:             "uatom",
			HostZoneId:        "cosmos",
			EpochNumber:       uint64(101),
			ClaimIsPending:    false,
		},
	} {
		s.App.StakeibcKeeper.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, *userRedemptionRecord)
	}

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
						Amount:                 sdkmath.NewInt(1200),
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
						Amount:                 sdkmath.NewInt(2000),
						Denom:                  "ustrd",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
					{
						Address:                "strideAddrUserA",
						Receiver:               "strideAddrUserA",
						Amount:                 sdkmath.NewInt(5000),
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
						Amount:                 sdkmath.NewInt(8500),
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
						Amount:                 sdkmath.NewInt(2000),
						Denom:                  "ustrd",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
						UnbondingEstimatedTime: "2024-01-11 00:00:00 +0000 UTC",
					},
					{
						Address:                "cosmosAddrUserA",
						Receiver:               "cosmosAddrUserA",
						Amount:                 sdkmath.NewInt(1200),
						Denom:                  "uatom",
						EpochNumber:            uint64(101),
						ClaimIsPending:         false,
						UnbondingEstimatedTime: "2024-01-27 00:00:00 +0000 UTC",
					},
					{
						Address:                "strideAddrUserA",
						Receiver:               "strideAddrUserA",
						Amount:                 sdkmath.NewInt(5000),
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
		s.Run(tc.desc, func() {
			response, err := s.App.StakeibcKeeper.AddressUnbondings(s.Ctx, tc.request)
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(
					tc.response,
					response,
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestEpochTrackerQuerySingle() {
	msgs := s.createNEpochTracker(2)
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
		s.Run(tc.desc, func() {
			response, err := s.App.StakeibcKeeper.EpochTracker(s.Ctx, tc.request)
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(
					tc.response,
					response,
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAllEpochTrackerQuery() {
	msgs := s.createNEpochTracker(5)

	resp, err := s.App.StakeibcKeeper.EpochTrackerAll(s.Ctx, &types.QueryAllEpochTrackerRequest{})
	s.Require().NoError(err)
	s.Require().Len(resp.EpochTracker, 5)
	s.Require().Subset(
		msgs,
		resp.EpochTracker,
	)
}

func (s *KeeperTestSuite) TestNextPacketSequenceQuery() {
	portId := "transfer"
	channelId := "channel-0"
	sequence := uint64(10)
	context := s.Ctx

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
