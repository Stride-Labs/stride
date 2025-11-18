package keeper_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v30/app/apptesting"
	"github.com/Stride-Labs/stride/v30/x/strdburner/keeper"
	"github.com/Stride-Labs/stride/v30/x/strdburner/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	logBuffer bytes.Buffer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Create a logger with accessible output
	logger := log.NewLogger(&s.logBuffer)
	s.Ctx = s.Ctx.WithLogger(logger)
}

func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.StrdBurnerKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestGetStrdBurnerAddress() {
	address := s.App.StrdBurnerKeeper.GetStrdBurnerAddress()
	require.NotNil(s.T(), address)
	require.Equal(s.T(), types.ModuleName, s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName).GetName())
}

func (s *KeeperTestSuite) TestSetAndGetProtocolStrdBurned() {
	// Test initial state (should be zero)
	initialAmount := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), initialAmount)

	// Clear any potential existing value to explicitly test nil case
	store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))
	store.Delete([]byte(types.ProtocolStrdBurnedKey))

	// Test getting value when none exists (should return zero)
	nilAmount := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), nilAmount)

	// Test setting and getting a value
	testAmount := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.SetProtocolStrdBurned(s.Ctx, testAmount)

	storedAmount := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), testAmount, storedAmount)

	// Test updating the value
	newAmount := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.SetProtocolStrdBurned(s.Ctx, newAmount)

	updatedAmount := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), newAmount, updatedAmount)

	// Confirm other burn amounts
	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), newAmount, totalBurned)

	userBurned := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), userBurned)
}

func (s *KeeperTestSuite) TestSetAndGetUserStrdBurned() {
	// Test initial state (should be zero)
	initialAmount := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), initialAmount)

	// Clear any potential existing value to explicitly test nil case
	store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))
	store.Delete([]byte(types.TotalUserStrdBurnedKey))

	// Test getting value when none exists (should return zero)
	nilAmount := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), nilAmount)

	// Test setting and getting a value
	testAmount := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.SetTotalUserStrdBurned(s.Ctx, testAmount)

	storedAmount := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	require.Equal(s.T(), testAmount, storedAmount)

	// Test updating the value
	newAmount := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.SetTotalUserStrdBurned(s.Ctx, newAmount)

	updatedAmount := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	require.Equal(s.T(), newAmount, updatedAmount)

	// Confirm other burn amounts
	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), newAmount, totalBurned)

	protocolBurned := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), protocolBurned)
}

func (s *KeeperTestSuite) TestSetAndGetStrdBurnedByAddress() {
	acc1 := s.TestAccs[0]
	acc2 := s.TestAccs[1]
	acc3 := s.TestAccs[2]

	// Test initial state (should be zero)
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1))
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2))
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc3))

	// Test setting and getting a value for user 1
	testAmount1 := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc1, testAmount1)

	require.Equal(s.T(), testAmount1, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1))
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2))
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc3))

	// Test setting and getting a value for user 2
	testAmount2 := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc2, testAmount2)

	require.Equal(s.T(), testAmount1, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1))
	require.Equal(s.T(), testAmount2, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2))
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc3))

	// Test setting and getting a value for user 3
	testAmount3 := sdkmath.NewInt(3000)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc3, testAmount3)

	require.Equal(s.T(), testAmount1, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1))
	require.Equal(s.T(), testAmount2, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2))
	require.Equal(s.T(), testAmount3, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc3))

	// Test updating the value
	newAmount1 := sdkmath.NewInt(4000)
	newAmount2 := sdkmath.NewInt(5000)
	newAmount3 := sdkmath.NewInt(6000)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc1, newAmount1)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc2, newAmount2)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc3, newAmount3)

	require.Equal(s.T(), newAmount1, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1))
	require.Equal(s.T(), newAmount2, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2))
	require.Equal(s.T(), newAmount3, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc3))
}

func (s *KeeperTestSuite) TestIncrementProtocolStrdBurned() {
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx))

	incrementAmount1 := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.IncrementProtocolStrdBurned(s.Ctx, incrementAmount1)
	require.Equal(s.T(), incrementAmount1, s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx))

	incrementAmount2 := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.IncrementProtocolStrdBurned(s.Ctx, incrementAmount2)
	require.Equal(s.T(), incrementAmount1.Add(incrementAmount2), s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx))
}

func (s *KeeperTestSuite) TestIncrementTotalUserStrdBurned() {
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx))

	incrementAmount1 := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.IncrementTotalUserStrdBurned(s.Ctx, incrementAmount1)
	require.Equal(s.T(), incrementAmount1, s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx))

	incrementAmount2 := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.IncrementTotalUserStrdBurned(s.Ctx, incrementAmount2)
	require.Equal(s.T(), incrementAmount1.Add(incrementAmount2), s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx))
}

func (s *KeeperTestSuite) TestIncrementStrdBurnedByAddress() {
	address := s.TestAccs[0]
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address))

	incrementAmount1 := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.IncrementStrdBurnedByAddress(s.Ctx, address, incrementAmount1)
	require.Equal(s.T(), incrementAmount1, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address))

	incrementAmount2 := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.IncrementStrdBurnedByAddress(s.Ctx, address, incrementAmount2)
	require.Equal(s.T(), incrementAmount1.Add(incrementAmount2), s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, address))
}

func (s *KeeperTestSuite) TestGetAllStrdBurnedAcrossAddresses() {
	acc1, acc2 := s.TestAccs[0], s.TestAccs[1]

	amount1 := sdkmath.NewInt(1000)
	amount2 := sdkmath.NewInt(2000)

	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc1, amount1)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc2, amount2)

	burnedAccounts := s.App.StrdBurnerKeeper.GetAllStrdBurnedAcrossAddresses(s.Ctx)
	s.Require().Len(burnedAccounts, 2)

	addressToAmount := make(map[string]sdkmath.Int)
	for _, account := range burnedAccounts {
		addressToAmount[account.Address] = account.Amount
	}

	s.Require().Contains(addressToAmount, acc1.String(), "account 1 should be present")
	s.Require().Contains(addressToAmount, acc2.String(), "account 2 should be present")

	s.Require().Equal(amount1, addressToAmount[acc1.String()], "account 1 amount")
	s.Require().Equal(amount2, addressToAmount[acc2.String()], "account 2 amount")
}

func (s *KeeperTestSuite) TestSetAndGetLinkedAddress() {
	acc1 := s.TestAccs[0]
	acc2 := s.TestAccs[1]
	acc3 := s.TestAccs[2]

	// Test initial state (should be empty)
	require.Equal(s.T(), "", s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc1))
	require.Equal(s.T(), "", s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc2))
	require.Equal(s.T(), "", s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc3))

	// Test setting and getting a value for user 1
	linkedAddress1 := "0x1"
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc1, linkedAddress1)

	require.Equal(s.T(), linkedAddress1, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc1))
	require.Equal(s.T(), "", s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc2))
	require.Equal(s.T(), "", s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc3))

	// Test setting and getting a value for user 2
	linkedAddress2 := "0x2"
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc2, linkedAddress2)

	require.Equal(s.T(), linkedAddress1, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc1))
	require.Equal(s.T(), linkedAddress2, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc2))
	require.Equal(s.T(), "", s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc3))

	// Test setting and getting a value for user 3
	linkedAddress3 := "0x3"
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc3, linkedAddress3)

	require.Equal(s.T(), linkedAddress1, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc1))
	require.Equal(s.T(), linkedAddress2, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc2))
	require.Equal(s.T(), linkedAddress3, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc3))

	// Test updating the value
	newLinkedAddress1 := "0x4"
	newLinkedAddress2 := "0x5"
	newLinkedAddress3 := "0x6"
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc1, newLinkedAddress1)
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc2, newLinkedAddress2)
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc3, newLinkedAddress3)

	require.Equal(s.T(), newLinkedAddress1, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc1))
	require.Equal(s.T(), newLinkedAddress2, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc2))
	require.Equal(s.T(), newLinkedAddress3, s.App.StrdBurnerKeeper.GetLinkedAddress(s.Ctx, acc3))
}

func (s *KeeperTestSuite) TestGetAllLinkedAddresses() {
	acc1, acc2 := s.TestAccs[0], s.TestAccs[1]

	linkedAddress1 := "0x1"
	linkedAddress2 := "0x2"

	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc1, linkedAddress1)
	s.App.StrdBurnerKeeper.SetLinkedAddress(s.Ctx, acc2, linkedAddress2)

	burnedAccounts := s.App.StrdBurnerKeeper.GetAllLinkedAddresses(s.Ctx)
	s.Require().Len(burnedAccounts, 2)

	strideToLinkedAddress := make(map[string]string)
	for _, account := range burnedAccounts {
		strideToLinkedAddress[account.StrideAddress] = account.LinkedAddress
	}

	s.Require().Contains(strideToLinkedAddress, acc1.String(), "account 1 should be present")
	s.Require().Contains(strideToLinkedAddress, acc2.String(), "account 2 should be present")

	s.Require().Equal(linkedAddress1, strideToLinkedAddress[acc1.String()], "account 1 linked address")
	s.Require().Equal(linkedAddress2, strideToLinkedAddress[acc2.String()], "account 2 linked address")
}

func (s *KeeperTestSuite) TestLogger() {
	logger := s.App.StrdBurnerKeeper.Logger(s.Ctx)
	require.NotNil(s.T(), logger)
}
