package v33_test

import (
	"encoding/hex"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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
		s.Require().Contains(val.Metadata.OperatorAddress, "stridevaloper")
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
