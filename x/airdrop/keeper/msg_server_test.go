package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v30/x/airdrop/types"
)

func (s *KeeperTestSuite) TestCreateAirdrop() {
	// Create a new airdrop
	msg := types.MsgCreateAirdrop{
		AirdropId:             AirdropId,
		DistributionStartDate: &DistributionStartDate,
		DistributionEndDate:   &DistributionEndDate,
		ClawbackDate:          &ClawbackDate,
		ClaimTypeDeadlineDate: &DeadlineDate,
		EarlyClaimPenalty:     sdkmath.LegacyMustNewDecFromStr("0.5"),
		DistributorAddress:    "distributor",
		AllocatorAddress:      "allocator",
		LinkerAddress:         "linker",
	}
	_, err := s.GetMsgServer().CreateAirdrop(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when creating airdrop")

	// Confirm the airdrop was a created
	airdrop := s.MustGetAirdrop(AirdropId)
	s.Require().Equal(AirdropId, airdrop.Id, "airdrop ID")
	s.Require().Equal(msg.DistributionStartDate, airdrop.DistributionStartDate, "distribution start date")
	s.Require().Equal(msg.DistributionEndDate, airdrop.DistributionEndDate, "distribution end date")
	s.Require().Equal(msg.ClawbackDate, airdrop.ClawbackDate, "clawback date")
	s.Require().Equal(msg.ClaimTypeDeadlineDate, airdrop.ClaimTypeDeadlineDate, "deadline date")
	s.Require().Equal(msg.EarlyClaimPenalty, airdrop.EarlyClaimPenalty, "early claim penalty")
	s.Require().Equal(msg.DistributorAddress, airdrop.DistributorAddress, "distributor address")
	s.Require().Equal(msg.AllocatorAddress, airdrop.AllocatorAddress, "allocator address")
	s.Require().Equal(msg.LinkerAddress, airdrop.LinkerAddress, "linker address")

	// Attempt to create it again, it should fail
	_, err = s.GetMsgServer().CreateAirdrop(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrAirdropAlreadyExists)
}

func (s *KeeperTestSuite) TestUpdateAirdrop() {
	// Create an airdrop
	s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
		Id: AirdropId,
	})

	// Update the airdrop
	msg := types.MsgUpdateAirdrop{
		AirdropId:             AirdropId,
		DistributionStartDate: &DistributionStartDate,
		DistributionEndDate:   &DistributionEndDate,
		ClawbackDate:          &ClawbackDate,
		ClaimTypeDeadlineDate: &DeadlineDate,
		EarlyClaimPenalty:     sdkmath.LegacyMustNewDecFromStr("0.8"),
		DistributorAddress:    "distributor2",
		AllocatorAddress:      "allocator2",
		LinkerAddress:         "linker2",
	}
	_, err := s.GetMsgServer().UpdateAirdrop(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when updating airdrop")

	// Confirm the airdrop was updated
	airdrop := s.MustGetAirdrop(AirdropId)
	s.Require().Equal(AirdropId, airdrop.Id, "airdrop ID")
	s.Require().Equal(msg.DistributionStartDate, airdrop.DistributionStartDate, "distribution start date")
	s.Require().Equal(msg.DistributionEndDate, airdrop.DistributionEndDate, "distribution end date")
	s.Require().Equal(msg.ClawbackDate, airdrop.ClawbackDate, "clawback date")
	s.Require().Equal(msg.ClaimTypeDeadlineDate, airdrop.ClaimTypeDeadlineDate, "deadline date")
	s.Require().Equal(msg.EarlyClaimPenalty, airdrop.EarlyClaimPenalty, "early claim penalty")
	s.Require().Equal(msg.DistributorAddress, airdrop.DistributorAddress, "distributor address")
	s.Require().Equal(msg.AllocatorAddress, airdrop.AllocatorAddress, "allocator address")
	s.Require().Equal(msg.LinkerAddress, airdrop.LinkerAddress, "linker address")

	// Remove the airdrop and try it again, it should error saying the airdrop doesn't exist
	s.App.AirdropKeeper.RemoveAirdrop(s.Ctx, AirdropId)
	_, err = s.GetMsgServer().UpdateAirdrop(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrAirdropNotFound)
}

func (s *KeeperTestSuite) TestAddAllocations() {
	// Create airdrop that's 4 days long
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)

	s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
		Id:                    AirdropId,
		DistributionStartDate: &startDate,
		DistributionEndDate:   &endDate,
	})

	// Submit a message to add allocations
	msg := types.MsgAddAllocations{
		AirdropId: AirdropId,
		Allocations: []types.RawAllocation{
			{
				UserAddress: "user-1",
				Allocations: []sdkmath.Int{
					sdkmath.NewInt(1),
					sdkmath.NewInt(2),
					sdkmath.NewInt(3),
					sdkmath.NewInt(4),
				},
			},
			{
				UserAddress: "user-2",
				Allocations: []sdkmath.Int{
					sdkmath.NewInt(4),
					sdkmath.NewInt(5),
					sdkmath.NewInt(6),
					sdkmath.NewInt(7),
				},
			},
		},
	}
	_, err := s.GetMsgServer().AddAllocations(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when adding allocations")

	// Confirm allocations were created for each user
	for _, expectedAllocation := range msg.Allocations {
		userAllocation, found := s.App.AirdropKeeper.GetUserAllocation(s.Ctx, AirdropId, expectedAllocation.UserAddress)
		s.Require().True(found, "user allocation %s should have been created", expectedAllocation.UserAddress)

		s.Require().Equal(AirdropId, userAllocation.AirdropId, "airdrop ID")
		s.Require().Equal(expectedAllocation.UserAddress, userAllocation.Address, "user address")
		s.Require().Equal(expectedAllocation.Allocations, userAllocation.Allocations, "allocations")
		s.Require().Equal(sdkmath.ZeroInt(), userAllocation.Claimed, "claimed")
		s.Require().Equal(sdkmath.ZeroInt(), userAllocation.Forfeited, "forfeited")
	}

	// Attempt to create the allocations again, it should error since the allocations already exist
	_, err = s.GetMsgServer().AddAllocations(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrUserAllocationAlreadyExists)

	// Try to add allocations with an incorrect number of elements, it should error
	invalidMsg := types.MsgAddAllocations{
		AirdropId: AirdropId,
		Allocations: []types.RawAllocation{
			{
				UserAddress: "user-3",
				Allocations: []sdkmath.Int{
					sdkmath.NewInt(1),
					sdkmath.NewInt(2),
					sdkmath.NewInt(3), // one less than above allocations
				},
			},
		},
	}
	_, err = s.GetMsgServer().AddAllocations(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorIs(err, types.ErrInvalidAllocationListLength)

	// Try to add the allocation with a nonadmin address, it should fail
	invalidMsg = msg
	invalidMsg.Admin = "different"
	_, err = s.GetMsgServer().AddAllocations(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorIs(err, types.ErrInvalidAdminAddress)

	// Remove the airdrop and try it again, it should error saying the airdrop doesn't exist
	s.App.AirdropKeeper.RemoveAirdrop(s.Ctx, AirdropId)
	_, err = s.GetMsgServer().AddAllocations(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrAirdropNotFound)
}

func (s *KeeperTestSuite) TestUpdateUserAllocation() {
	allocator := s.TestAccs[0]

	initialAllocations := []sdkmath.Int{sdkmath.NewInt(1), sdkmath.NewInt(2)}
	updatedAllocations := []sdkmath.Int{sdkmath.NewInt(3), sdkmath.NewInt(4)}

	// Create an airdrop and user allocation
	s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
		Id:               AirdropId,
		AllocatorAddress: allocator.String(),
	})
	s.App.AirdropKeeper.SetUserAllocation(s.Ctx, types.UserAllocation{
		AirdropId:   AirdropId,
		Address:     UserAddress,
		Allocations: initialAllocations,
	})

	// Update the allocations
	msg := types.MsgUpdateUserAllocation{
		Admin:       allocator.String(),
		AirdropId:   AirdropId,
		UserAddress: UserAddress,
		Allocations: updatedAllocations,
	}
	_, err := s.GetMsgServer().UpdateUserAllocation(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when updating allocation")

	// Try to update again to a different allocation length, it should fail
	invalidMsg := msg
	invalidMsg.Allocations = updatedAllocations[1:] // trimmed first element
	_, err = s.GetMsgServer().UpdateUserAllocation(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorIs(err, types.ErrInvalidAllocationListLength)

	// Remove the user allocation and try again, it should error since it's not found
	s.App.AirdropKeeper.RemoveUserAllocation(s.Ctx, AirdropId, UserAddress)
	_, err = s.GetMsgServer().UpdateUserAllocation(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrUserAllocationNotFound)

	// Try to update the allocation with a nonadmin address, it should fail
	invalidMsg = msg
	invalidMsg.Admin = "different"
	_, err = s.GetMsgServer().UpdateUserAllocation(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorIs(err, types.ErrInvalidAdminAddress)

	// Remove the airdrop and try again, it should also error
	s.App.AirdropKeeper.RemoveAirdrop(s.Ctx, AirdropId)
	_, err = s.GetMsgServer().UpdateUserAllocation(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrAirdropNotFound)
}

func (s *KeeperTestSuite) TestMsgLinkAddresses() {
	// Most test cases are covered in the keeper function
	// This is meant mainly to test the happy path and admin address enforcement
	strideAddress := "stride"
	hostAddress := "host"
	linkerAddress := "linker"

	// Create the initial airdrop (it only has to exist)
	s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
		Id:            AirdropId,
		LinkerAddress: linkerAddress,
	})

	// Create the stride and host allocations
	strideUserAllocation := types.UserAllocation{
		AirdropId:   AirdropId,
		Address:     strideAddress,
		Allocations: []sdkmath.Int{sdkmath.NewInt(10), sdkmath.NewInt(10), sdkmath.NewInt(10)},
		Forfeited:   sdkmath.ZeroInt(),
		Claimed:     sdkmath.NewInt(10),
	}
	hostUserAllocation := types.UserAllocation{
		AirdropId:   AirdropId,
		Address:     hostAddress,
		Allocations: []sdkmath.Int{sdkmath.NewInt(10), sdkmath.NewInt(10), sdkmath.NewInt(10)},
		Forfeited:   sdkmath.ZeroInt(),
		Claimed:     sdkmath.ZeroInt(),
	}
	s.App.AirdropKeeper.SetUserAllocation(s.Ctx, strideUserAllocation)
	s.App.AirdropKeeper.SetUserAllocation(s.Ctx, hostUserAllocation)

	expectedUpdatedUserAllocation := types.UserAllocation{
		AirdropId:   AirdropId,
		Address:     strideAddress,
		Claimed:     sdkmath.NewInt(10),
		Forfeited:   sdkmath.ZeroInt(),
		Allocations: []sdkmath.Int{sdkmath.NewInt(20), sdkmath.NewInt(20), sdkmath.NewInt(20)},
	}

	// Call link
	msg := types.MsgLinkAddresses{
		Admin:         linkerAddress,
		AirdropId:     AirdropId,
		StrideAddress: strideAddress,
		HostAddress:   hostAddress,
	}
	_, err := s.GetMsgServer().LinkAddresses(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected during linking")

	// Check that the stride allocation was updated
	actualUserAllocation := s.MustGetUserAllocation(AirdropId, strideAddress)
	s.Require().Equal(expectedUpdatedUserAllocation, actualUserAllocation, "updated user allocation")

	// Check that the host user was removed
	_, hostFound := s.App.AirdropKeeper.GetUserAllocation(s.Ctx, AirdropId, hostAddress)
	s.Require().False(hostFound, "host user allocation should have been removed")

	// Attempt to call it again with a non-admin address, it should fail
	invalidMsg := msg
	invalidMsg.Admin = "different"
	_, err = s.GetMsgServer().LinkAddresses(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorIs(err, types.ErrInvalidAdminAddress)

	// Remove the airdrop and try it again, it should error even sooner
	s.App.AirdropKeeper.RemoveAirdrop(s.Ctx, AirdropId)
	_, err = s.GetMsgServer().LinkAddresses(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorIs(err, types.ErrAirdropNotFound)
}
