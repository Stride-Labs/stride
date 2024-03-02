package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	epochtypes "github.com/Stride-Labs/stride/v18/x/epochs/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

const nanosecondsInDay = 86400000000000

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) ModuleAddress(goCtx context.Context, req *types.QueryModuleAddressRequest) (*types.QueryModuleAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := k.AccountKeeper.GetModuleAccount(ctx, req.Name).GetAddress().String()

	return &types.QueryModuleAddressResponse{Addr: addr}, nil
}

func (k Keeper) HostZoneAll(c context.Context, req *types.QueryAllHostZoneRequest) (*types.QueryAllHostZoneResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var hostZones []types.HostZone
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	hostZoneStore := prefix.NewStore(store, types.KeyPrefix(types.HostZoneKey))

	pageRes, err := query.Paginate(hostZoneStore, req.Pagination, func(key []byte, value []byte) error {
		var hostZone types.HostZone
		if err := k.cdc.Unmarshal(value, &hostZone); err != nil {
			return err
		}

		hostZones = append(hostZones, hostZone)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllHostZoneResponse{HostZone: hostZones, Pagination: pageRes}, nil
}

func (k Keeper) HostZone(c context.Context, req *types.QueryGetHostZoneRequest) (*types.QueryGetHostZoneResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	hostZone, found := k.GetHostZone(ctx, req.ChainId)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetHostZoneResponse{HostZone: hostZone}, nil
}

func (k Keeper) Validators(c context.Context, req *types.QueryGetValidatorsRequest) (*types.QueryGetValidatorsResponse, error) {
	if req == nil || req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	hostZone, found := k.GetHostZone(ctx, req.ChainId)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetValidatorsResponse{Validators: hostZone.Validators}, nil
}

func (k Keeper) AddressUnbondings(c context.Context, req *types.QueryAddressUnbondings) (*types.QueryAddressUnbondingsResponse, error) {
	// The function queries all the unbondings associated with Stride addresses.
	// This should provide more visiblity into the unbonding process for a user.

	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// The address field can either be a single address or several comma separated
	addresses := strings.Split(req.Address, ",")

	addressUnbondings := []types.AddressUnbonding{}

	// get the relevant day
	dayEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.DAY_EPOCH)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}
	currentDay := dayEpochTracker.EpochNumber

	epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		for _, hostZoneUnbonding := range epochUnbondingRecord.GetHostZoneUnbondings() {
			for _, userRedemptionRecordId := range hostZoneUnbonding.GetUserRedemptionRecords() {
				userRedemptionRecordComponents := strings.Split(userRedemptionRecordId, ".")
				if len(userRedemptionRecordComponents) != 3 {
					k.Logger(ctx).Error(fmt.Sprintf("invalid user redemption record id %s", userRedemptionRecordId))
					continue
				}
				userRedemptionRecordAddress := userRedemptionRecordComponents[2]

				// Check if the userRedemptionRecordAddress is one targeted by the address(es) in the query
				targetAddress := false
				for _, address := range addresses {
					if userRedemptionRecordAddress == strings.TrimSpace(address) {
						targetAddress = true
						break
					}
				}
				if targetAddress {
					userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
					if !found {
						continue // the record has already been claimed
					}

					// get the anticipated unbonding time
					unbondingTime := hostZoneUnbonding.UnbondingTime
					if unbondingTime == 0 {
						hostZone, found := k.GetHostZone(ctx, hostZoneUnbonding.HostZoneId)
						if !found {
							return nil, sdkerrors.ErrKeyNotFound
						}
						unbondingFrequency := hostZone.GetUnbondingFrequency()
						daysUntilUnbonding := unbondingFrequency - (currentDay % unbondingFrequency)
						unbondingStartTime := dayEpochTracker.NextEpochStartTime + ((daysUntilUnbonding - 1) * nanosecondsInDay)
						unbondingDurationEstimate := (unbondingFrequency - 1) * 7
						unbondingTime = unbondingStartTime + (unbondingDurationEstimate * nanosecondsInDay)
					}
					unbondingTime = unbondingTime + nanosecondsInDay
					unbondingTimeStr := time.Unix(0, int64(unbondingTime)).UTC().String()

					addressUnbonding := types.AddressUnbonding{
						Address:                userRedemptionRecordAddress,
						Receiver:               userRedemptionRecord.Receiver,
						UnbondingEstimatedTime: unbondingTimeStr,
						Amount:                 userRedemptionRecord.NativeTokenAmount,
						Denom:                  userRedemptionRecord.Denom,
						ClaimIsPending:         userRedemptionRecord.ClaimIsPending,
						EpochNumber:            userRedemptionRecord.EpochNumber,
					}
					addressUnbondings = append(addressUnbondings, addressUnbonding)
				}
			}
		}
	}

	return &types.QueryAddressUnbondingsResponse{AddressUnbondings: addressUnbondings}, nil
}

func (k Keeper) AllTradeRoutes(c context.Context, req *types.QueryAllTradeRoutes) (*types.QueryAllTradeRoutesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	routes := k.GetAllTradeRoutes(ctx)

	return &types.QueryAllTradeRoutesResponse{TradeRoutes: routes}, nil
}

// InterchainAccountFromAddress implements the Query/InterchainAccountFromAddress gRPC method
func (k Keeper) InterchainAccountFromAddress(goCtx context.Context, req *types.QueryInterchainAccountFromAddressRequest) (*types.QueryInterchainAccountFromAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	portID, err := icatypes.NewControllerPortID(req.Owner)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not find account: %s", err)
	}

	addr, found := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, req.ConnectionId, portID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no account found for portID %s", portID)
	}

	return types.NewQueryInterchainAccountResponse(addr), nil
}

func (k Keeper) NextPacketSequence(c context.Context, req *types.QueryGetNextPacketSequenceRequest) (*types.QueryGetNextPacketSequenceResponse, error) {
	if req == nil || req.ChannelId == "" || req.PortId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	sequence, found := k.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, req.PortId, req.ChannelId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "channel and port combination not found")
	}

	return &types.QueryGetNextPacketSequenceResponse{Sequence: sequence}, nil
}
