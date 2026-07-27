package v33_test

import (
	"encoding/hex"
	"testing"

	v4 "github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/migrations/v4"
	"github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/migrations/v4/legacy"
	packetforwardtypes "github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/types"
	consumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v33/app/apptesting"
	v33 "github.com/Stride-Labs/stride/v33/app/upgrades/v33"
	"github.com/Stride-Labs/stride/v33/utils"
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

	// Map each seeded consensus address to one of the real monikers in
	// utils.PoaValidatorSet so SnapshotValidatorsFromICS can complete the
	// hex_cons_addr → moniker → operator join.
	addrs := s.getSeededConsAddresses()
	s.Require().Len(addrs, len(utils.PoaValidatorSet))

	expectedOperators := make(map[string]string, len(addrs)) // hex_cons_addr → operator
	for i, addr := range addrs {
		moniker := utils.PoaValidatorSet[i].Moniker
		v33.ValidatorMonikers[hex.EncodeToString(addr)] = moniker
		expectedOperators[hex.EncodeToString(addr)] = utils.PoaValidatorSet[i].Operator
	}
	s.T().Cleanup(func() {
		for _, addr := range addrs {
			delete(v33.ValidatorMonikers, hex.EncodeToString(addr))
		}
	})

	poaValidators, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().NoError(err)
	s.Require().Len(poaValidators, 8)

	for _, val := range poaValidators {
		s.Require().NotNil(val.PubKey)
		s.Require().Equal(int64(100), val.Power)
		s.Require().NotNil(val.Metadata)

		// Recover the seeded hex_cons_addr from the validator's pubkey so we can
		// look up the moniker + operator we expected to be assigned to it.
		var pk cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(val.PubKey, &pk))
		hexAddr := hex.EncodeToString(pk.Address().Bytes())

		s.Require().Equal(v33.ValidatorMonikers[hexAddr], val.Metadata.Moniker)
		s.Require().Equal(expectedOperators[hexAddr], val.Metadata.OperatorAddress)
	}
}

func (s *HelpersTestSuite) TestSnapshotValidatorsFromICS_TooManyValidators() {
	s.seedConsumerValidators(9)

	_, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "expected at most 8 validators")
}

func (s *HelpersTestSuite) TestSnapshotValidatorsFromICS_MissingMoniker() {
	s.seedConsumerValidators(8)
	// Do NOT populate ValidatorMonikers — every validator hex_cons_addr lookup
	// will miss, which should now halt the upgrade rather than silently
	// produce empty monikers.

	_, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "no moniker in v33 validators.json")
}

func (s *HelpersTestSuite) TestSnapshotValidatorsFromICS_UnknownMoniker() {
	s.seedConsumerValidators(8)

	// Populate monikers but use a value that does NOT appear in
	// utils.PoaValidatorSet — this catches drift between the two sources of
	// truth.
	addrs := s.getSeededConsAddresses()
	for _, addr := range addrs {
		v33.ValidatorMonikers[hex.EncodeToString(addr)] = "moniker-not-in-poa-set"
	}
	s.T().Cleanup(func() {
		for _, addr := range addrs {
			delete(v33.ValidatorMonikers, hex.EncodeToString(addr))
		}
	})

	_, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "no entry in utils.PoaValidatorSet")
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

	// SnapshotValidatorsFromICS now requires every CCValidator to map to a
	// real moniker + operator. Wire that up before the call.
	addrs := s.getSeededConsAddresses()
	for i, addr := range addrs {
		v33.ValidatorMonikers[hex.EncodeToString(addr)] = utils.PoaValidatorSet[i].Moniker
	}
	s.T().Cleanup(func() {
		for _, addr := range addrs {
			delete(v33.ValidatorMonikers, hex.EncodeToString(addr))
		}
	})

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

// legacyInFlightPacket returns a well-formed in-flight packet, the shape all 25 legitimate
// mainnet records have.
func legacyInFlightPacket() legacy.InFlightPacket {
	return legacy.InFlightPacket{
		OriginalSenderAddress:  "osmo1m8wg4vxkefhs374qxmmqpyusgz289wmulex5qdwpfx7jnrxzer5s9cv83q",
		RefundChannelId:        "channel-5",
		RefundPortId:           "transfer",
		PacketSrcChannelId:     "channel-326",
		PacketSrcPortId:        "transfer",
		PacketTimeoutTimestamp: 1739373116478544000,
		PacketTimeoutHeight:    "0-0",
		PacketData:             []byte(`{"amount":"1"}`),
		RefundSequence:         532899,
		RetriesRemaining:       2,
		Timeout:                600000000000,
	}
}

// degenerateInFlightPacket mirrors the single junk record on Stride mainnet: stored under key
// 0x00, no refund routing, and an empty timeout height that MustParseHeight panics on.
func degenerateInFlightPacket() legacy.InFlightPacket {
	return legacy.InFlightPacket{
		OriginalSenderAddress: "0",
		PacketTimeoutHeight:   "",
	}
}

func (s *HelpersTestSuite) setInFlightPacket(key []byte, packet legacy.InFlightPacket) {
	store := s.Ctx.KVStore(s.App.GetKey(packetforwardtypes.StoreKey))
	store.Set(key, s.App.AppCodec().MustMarshal(&packet))
}

func (s *HelpersTestSuite) inFlightPacketExists(key []byte) bool {
	store := s.Ctx.KVStore(s.App.GetKey(packetforwardtypes.StoreKey))
	return store.Has(key)
}

func (s *HelpersTestSuite) TestPruneMalformedInFlightPackets_RemovesDegenerateRecord() {
	validKey := []byte("channel-0/transfer/161399")
	degenerateKey := []byte{0}

	s.setInFlightPacket(validKey, legacyInFlightPacket())
	s.setInFlightPacket(degenerateKey, degenerateInFlightPacket())

	pruned, err := v33.PruneMalformedInFlightPackets(s.Ctx, s.App.AppCodec(), s.App.GetKey(packetforwardtypes.StoreKey))
	s.Require().NoError(err)
	s.Require().Equal(1, pruned, "only the degenerate record should be pruned")

	s.Require().False(s.inFlightPacketExists(degenerateKey), "degenerate record should be deleted")
	s.Require().True(s.inFlightPacketExists(validKey), "legitimate record must be left alone")
}

func (s *HelpersTestSuite) TestPruneMalformedInFlightPackets_NoopWhenAllValid() {
	s.setInFlightPacket([]byte("channel-0/transfer/1"), legacyInFlightPacket())
	s.setInFlightPacket([]byte("channel-0/transfer/2"), legacyInFlightPacket())

	pruned, err := v33.PruneMalformedInFlightPackets(s.Ctx, s.App.AppCodec(), s.App.GetKey(packetforwardtypes.StoreKey))
	s.Require().NoError(err)
	s.Require().Zero(pruned)
}

// A packet that cannot be parsed but still carries refund routing represents real user funds.
// Deleting it would silently abandon the refund, so the upgrade must halt instead.
func (s *HelpersTestSuite) TestPruneMalformedInFlightPackets_RefusesRefundablePacket() {
	refundableKey := []byte("channel-0/transfer/999")

	refundable := legacyInFlightPacket()
	refundable.PacketTimeoutHeight = "not-a-height"
	s.setInFlightPacket(refundableKey, refundable)

	_, err := v33.PruneMalformedInFlightPackets(s.Ctx, s.App.AppCodec(), s.App.GetKey(packetforwardtypes.StoreKey))
	s.Require().ErrorContains(err, "refusing to drop a refundable packet")
	s.Require().True(s.inFlightPacketExists(refundableKey), "refundable record must be preserved")
}

// The regression that matters: ibc-go's PFM 3→4 migration panics on the degenerate record, which
// would halt every mainnet validator at the upgrade height. Pruning first makes it succeed.
func (s *HelpersTestSuite) TestPruneMalformedInFlightPackets_UnblocksPFMMigration() {
	storeKey := s.App.GetKey(packetforwardtypes.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	cdc := s.App.AppCodec()

	s.setInFlightPacket([]byte("channel-0/transfer/161399"), legacyInFlightPacket())
	s.setInFlightPacket([]byte{0}, degenerateInFlightPacket())

	s.Require().Panics(func() {
		_ = v4.Migrate(s.Ctx, storeService, cdc)
	}, "migration is expected to panic on the degenerate record without the prune")

	pruned, err := v33.PruneMalformedInFlightPackets(s.Ctx, cdc, storeKey)
	s.Require().NoError(err)
	s.Require().Equal(1, pruned)

	s.Require().NoError(v4.Migrate(s.Ctx, storeService, cdc), "migration should succeed after pruning")
}
