package v19_test

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v18/app/apptesting"
	v19 "github.com/Stride-Labs/stride/v18/app/upgrades/v19"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	// Run through upgrade
	s.ConfirmUpgradeSucceededs("v19", dummyUpgradeHeight)

	// Check state after upgrade
	s.CheckWasmPerms()
}

func (s *UpgradeTestSuite) CheckWasmPerms() {
	wasmParams := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, wasmParams.CodeUploadAccess.Permission, "upload permission")
	s.Require().Equal(v19.WasmAdmin, wasmParams.CodeUploadAccess.Addresses[0], "upload address")
	s.Require().Equal(wasmtypes.AccessTypeNobody, wasmParams.InstantiateDefaultPermission, "instantiate permission")
}
