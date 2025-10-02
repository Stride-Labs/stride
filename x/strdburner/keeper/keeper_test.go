package keeper_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
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

	userBurned := s.App.StrdBurnerKeeper.GetUserStrdBurned(s.Ctx)
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

	userBurned := s.App.StrdBurnerKeeper.GetUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), userBurned)
}

func (s *KeeperTestSuite) TestSetAndGetUserStrdBurned() {
	// Test initial state (should be zero)
	initialAmount := s.App.StrdBurnerKeeper.GetUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), initialAmount)

	// Clear any potential existing value to explicitly test nil case
	store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))
	store.Delete([]byte(types.UserStrdBurnedKey))

	// Test getting value when none exists (should return zero)
	nilAmount := s.App.StrdBurnerKeeper.GetUserStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), nilAmount)

	// Test setting and getting a value
	testAmount := sdkmath.NewInt(1000)
	s.App.StrdBurnerKeeper.SetUserStrdBurned(s.Ctx, testAmount)

	storedAmount := s.App.StrdBurnerKeeper.GetUserStrdBurned(s.Ctx)
	require.Equal(s.T(), testAmount, storedAmount)

	// Test updating the value
	newAmount := sdkmath.NewInt(2000)
	s.App.StrdBurnerKeeper.SetUserStrdBurned(s.Ctx, newAmount)

	updatedAmount := s.App.StrdBurnerKeeper.GetUserStrdBurned(s.Ctx)
	require.Equal(s.T(), newAmount, updatedAmount)

	// Confirm other burn amounts are 0
	totalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), totalBurned)

	protocolBurned := s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx)
	require.Equal(s.T(), sdkmath.ZeroInt(), protocolBurned)
}

func (s *KeeperTestSuite) TestLogger() {
	logger := s.App.StrdBurnerKeeper.Logger(s.Ctx)
	require.NotNil(s.T(), logger)
}
