package v33_test

import (
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	consumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
)

type HelpersTestSuite struct {
	apptesting.AppTestHelper
}

func (s *HelpersTestSuite) SetupTest() {
	s.Setup()
}

func TestHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(HelpersTestSuite))
}

func (s *HelpersTestSuite) TestSnapshotValidatorsFromICS_HappyPath() {
	s.seedConsumerValidators(8)

	addrs := s.getSeededConsAddresses()
	for i, addr := range addrs {
		v33.ValidatorMonikers[hex.EncodeToString(addr)] = fmtMoniker(i)
	}
	s.T().Cleanup(func() {
		for _, addr := range addrs {
			delete(v33.ValidatorMonikers, hex.EncodeToString(addr))
		}
	})

	poaValidators, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().NoError(err)
	s.Require().Len(poaValidators, 8)

	for i, val := range poaValidators {
		s.Require().NotNil(val.PubKey)
		s.Require().Equal(int64(100), val.Power)
		s.Require().NotNil(val.Metadata)
		s.Require().Equal(fmtMoniker(i), val.Metadata.Moniker)
		s.Require().Contains(val.Metadata.OperatorAddress, "stride1")
	}
}

func (s *HelpersTestSuite) TestSnapshotValidatorsFromICS_WrongCount() {
	s.seedConsumerValidators(7) // expecting 8

	_, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "expected 8 validators")
}

func (s *HelpersTestSuite) TestSnapshotValidatorsFromICS_UnknownMoniker() {
	s.seedConsumerValidators(8)
	// Do NOT populate ValidatorMonikers — every validator should fall back to ""

	poaValidators, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().NoError(err)
	for _, val := range poaValidators {
		s.Require().Equal("", val.Metadata.Moniker)
	}
}

func (s *HelpersTestSuite) TestSweepICSModuleAccounts_HappyPath() {
	consRedistributeAddr := s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerRedistributeName)
	consToProviderAddr := s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerToSendToProviderName)

	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	fundAmount := sdk.NewInt64Coin(bondDenom, 1_000_000)

	s.FundModuleAccount(consumertypes.ConsumerRedistributeName, fundAmount)
	s.FundModuleAccount(consumertypes.ConsumerToSendToProviderName, fundAmount)

	err = v33.SweepICSModuleAccounts(
		s.Ctx, s.App.AccountKeeper, s.App.BankKeeper, s.App.DistrKeeper,
	)
	s.Require().NoError(err)

	s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistributeAddr).IsZero())
	s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consToProviderAddr).IsZero())

	// Community pool should have received both amounts
	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	cpBalance := feePool.CommunityPool.AmountOf(bondDenom)
	s.Require().Equal(sdkmath.LegacyNewDec(2_000_000), cpBalance)
}

func (s *HelpersTestSuite) TestSweepICSModuleAccounts_EmptyAccounts() {
	// Don't fund anything — both accounts are empty; sweep should be a no-op
	err := v33.SweepICSModuleAccounts(
		s.Ctx, s.App.AccountKeeper, s.App.BankKeeper, s.App.DistrKeeper,
	)
	s.Require().NoError(err)
}

func (s *HelpersTestSuite) TestInitializePOA_HappyPath() {
	// Record how many POA validators exist before the operation. The test app
	// seeds one genesis validator so that POA's InitGenesis produces a non-empty
	// ValidatorUpdate for InitChain (required since ccvconsumer was removed from
	// the module manager). That bootstrap validator is test infrastructure and
	// should not affect the count of ICS-migrated validators.
	initialVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	initialCount := len(initialVals)

	s.seedConsumerValidators(8)

	poaVals, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().NoError(err)

	adminAddr := s.TestAccs[0].String()
	err = v33.InitializePOA(s.Ctx, s.App.AppCodec(), s.App.POAKeeper, adminAddr, poaVals)
	s.Require().NoError(err)

	storedVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	// InitializePOA should have added exactly 8 new validators from ICS.
	s.Require().Len(storedVals, initialCount+8)

	params, err := s.App.POAKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(adminAddr, params.Admin)
}

func (s *HelpersTestSuite) TestInitializePOA_InvalidAdmin() {
	err := v33.InitializePOA(s.Ctx, s.App.AppCodec(), s.App.POAKeeper, "not_a_bech32", []poatypes.Validator{})
	s.Require().Error(err)
}

// --- test helpers ---

func (s *HelpersTestSuite) seedConsumerValidators(count int) {
	// Clear any validators already present from genesis state before seeding.
	for _, existing := range s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx) {
		s.App.ConsumerKeeper.DeleteCCValidator(s.Ctx, existing.Address)
	}

	for i := 0; i < count; i++ {
		privKey := ed25519.GenPrivKeyFromSecret([]byte{byte(i + 1)})
		pubKey := privKey.PubKey()
		addr := pubKey.Address().Bytes()
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		s.Require().NoError(err)

		ccVal := consumertypes.CrossChainValidator{
			Address: addr,
			Power:   100,
			Pubkey:  pkAny,
		}
		s.App.ConsumerKeeper.SetCCValidator(s.Ctx, ccVal)
	}
}

func (s *HelpersTestSuite) getSeededConsAddresses() [][]byte {
	vals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
	out := make([][]byte, 0, len(vals))
	for _, v := range vals {
		out = append(out, v.Address)
	}
	return out
}

func fmtMoniker(i int) string {
	return "validator-" + string(rune('a'+i))
}
