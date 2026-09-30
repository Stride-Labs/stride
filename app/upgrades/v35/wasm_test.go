package v35_test

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/testdata"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
)

func (s *UpgradeTestSuite) TestSetWasmUploadAccessToGov() {
	deployKey := sdk.MustAccAddressFromBech32(v35.WasmDeployKey)
	params := s.App.WasmKeeper.GetParams(s.Ctx)
	params.CodeUploadAccess = wasmtypes.AccessTypeAnyOfAddresses.With(deployKey, apptesting.CreateRandomAccounts(1)[0])
	s.Require().NoError(s.App.WasmKeeper.SetParams(s.Ctx, params))

	err := v35.SetWasmUploadAccessToGov(s.Ctx, s.App.WasmKeeper)
	s.Require().NoError(err)

	after := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, after.CodeUploadAccess.Permission)
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, after.CodeUploadAccess.Addresses)
	s.Require().Equal(params.InstantiateDefaultPermission, after.InstantiateDefaultPermission, "instantiate permission untouched")
}

// storeAndInstantiateHackatom stores hackatom.wasm through the gov permission keeper (which
// bypasses upload access) and instantiates it with the given admin. Returns the contract address.
func (s *UpgradeTestSuite) storeAndInstantiateHackatom(admin sdk.AccAddress) sdk.AccAddress {
	govKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	creator := apptesting.CreateRandomAccounts(1)[0]

	codeId, _, err := govKeeper.Create(s.Ctx, creator, testdata.HackatomContractWasm(), nil)
	s.Require().NoError(err, "store hackatom")

	initMsg, err := json.Marshal(map[string]string{
		"verifier":    creator.String(),
		"beneficiary": creator.String(),
	})
	s.Require().NoError(err)

	contractAddr, _, err := govKeeper.Instantiate(s.Ctx, codeId, creator, admin, initMsg, "hackatom", sdk.NewCoins())
	s.Require().NoError(err, "instantiate hackatom")
	return contractAddr
}

func (s *UpgradeTestSuite) TestMoveDeployKeyContractAdminsToGov() {
	deployKey := sdk.MustAccAddressFromBech32(v35.WasmDeployKey)
	otherAdmin := apptesting.CreateRandomAccounts(1)[0]

	deployKeyContract := s.storeAndInstantiateHackatom(deployKey)
	otherAdminContract := s.storeAndInstantiateHackatom(otherAdmin)
	noAdminContract := s.storeAndInstantiateHackatom(nil)

	v35.MoveDeployKeyContractAdminsToGov(s.Ctx, s.App.WasmKeeper)

	gov := v35.GovModuleAddress().String()
	s.Require().Equal(gov, s.App.WasmKeeper.GetContractInfo(s.Ctx, deployKeyContract).Admin, "deploy-key contract moved to gov")
	s.Require().Equal(otherAdmin.String(), s.App.WasmKeeper.GetContractInfo(s.Ctx, otherAdminContract).Admin, "other admin untouched")
	s.Require().Equal("", s.App.WasmKeeper.GetContractInfo(s.Ctx, noAdminContract).Admin, "admin-less contract untouched")
}

func (s *UpgradeTestSuite) TestMoveDeployKeyContractAdminsToGov_NoContracts() {
	s.Require().NotPanics(func() { v35.MoveDeployKeyContractAdminsToGov(s.Ctx, s.App.WasmKeeper) })
}
