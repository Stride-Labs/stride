package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/x/airdrop/keeper"
	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

const (
	UserAddress = "address"
	AirdropId   = "airdrop"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.AirdropKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Helper function to get an airdrop and confirm there's no error
func (s *KeeperTestSuite) MustGetAirdrop(airdropId string) types.Airdrop {
	airdrop, found := s.App.AirdropKeeper.GetAirdrop(s.Ctx, airdropId)
	s.Require().True(found, "airdrop %s should have been found", airdropId)
	return airdrop
}

// Helper function to get a user allocation and confirm there's no error
func (s *KeeperTestSuite) MustGetUserAllocation(airdropId, address string) types.UserAllocation {
	userAllocation, found := s.App.AirdropKeeper.GetUserAllocation(s.Ctx, airdropId, address)
	s.Require().True(found, "user allocation for %s and %s should have been found", airdropId, address)
	return userAllocation
}
