package keeper_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	"github.com/Stride-Labs/stride/v28/x/strdburner/keeper"
	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
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

func (s *KeeperTestSuite) TestSetAndGetTotalStrdBurned() {
	// Test initial state (should be zero)
	initialAmount := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), initialAmount)

	// Clear any potential existing value to explicitly test nil case
	store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))
	store.Delete([]byte(types.TotalStrdBurnedKey))

	// Test getting value when none exists (should return zero)
	nilAmount := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), nilAmount)

	// Test setting and getting a value
	testAmount := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.SetTotalStrdBurned(s.Ctx, testAmount)

	storedAmount := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), testAmount, storedAmount)

	// Test updating the value
	newAmount := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.SetTotalStrdBurned(s.Ctx, newAmount)

	updatedAmount := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), newAmount, updatedAmount)

	// Confirm other burn amounts are 0
	protocolBurned := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), protocolBurned)

	userBurned := s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), userBurned)
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

	// Confirm other burn amounts are 0
	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), totalBurned)

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

	// Confirm other burn amounts are 0
	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), totalBurned)

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

func (s *KeeperTestSuite) TestIncrementTotalStrdBurned() {
	require.Equal(s.T(), sdkmath.ZeroInt(), s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx))

	incrementAmount1 := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.IncrementTotalStrdBurned(s.Ctx, incrementAmount1)
	require.Equal(s.T(), incrementAmount1, s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx))

	incrementAmount2 := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.IncrementTotalStrdBurned(s.Ctx, incrementAmount2)
	require.Equal(s.T(), incrementAmount1.Add(incrementAmount2), s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx))
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

func (s *KeeperTestSuite) TestLogger() {
	logger := s.App.StrdBurnerKeeper.Logger(s.Ctx)
	require.NotNil(s.T(), logger)
}
