package keeper_test

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	sdkmath "cosmossdk.io/math"

	epochtypes "github.com/Stride-Labs/stride/v11/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v11/x/records/types"
	"github.com/Stride-Labs/stride/v11/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v11/x/stakeibc/types"

	abci "github.com/cometbft/cometbft/abci/types"
)

func BenchmarkLiquidStaking(b *testing.B) {
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	msgs := []sdk.Msg{
		&types.MsgLiquidStake{
			Creator:   addr0.String(),
			Amount:    sdk.NewInt(1000000),
			HostDenom: Atom,
		},
	}
	benchmarkWrapper(b, msgs, priv0, false)

}

// benchmarkWrapper is a wrapper function for the benchmark tests. It sets up the suite, accepts the
// messages to be sent, and the expected number of trades. It then runs the benchmark and checks the
// number of trades after the post handler is run.
func benchmarkWrapper(b *testing.B, msgs []sdk.Msg, privKey cryptotypes.PrivKey, isHalted bool) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s, tx, _, _ := setUpBenchmarkSuite(msgs, privKey)

		b.StartTimer()

		txBytes, _ := s.App.GetTxConfig().TxEncoder()(tx)
		_ = s.App.BaseApp.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

		b.StopTimer()

		hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
		s.Require().True(found)

		if hz.Halted != isHalted {
			b.Fatalf("expected halted %v, got %v", isHalted, hz.Halted)
		}
	}
}

func (s *KeeperTestSuite) SetUpHostZoneAndAccount(privKey cryptotypes.PrivKey) (cryptotypes.PrivKey, sdk.AccAddress) {
	s.SetupTest()

	// Set up the app to the correct state to run the test
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// Init a new account and fund it with tokens for gas fees
	priv0 := privKey
	addr0 := sdk.AccAddress(privKey.PubKey().Address())
	acc0 := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr0)
	s.App.AccountKeeper.SetAccount(s.Ctx, acc0)
	s.FundAccount(addr0, sdk.NewInt64Coin(IbcAtom, 10_000_000))
	s.FundAccount(addr0, sdk.NewCoin("ustrd", sdk.NewInt(1000000)))

	// Init a new host zone account and fund it with tokens for gas fees
	zoneAddress := types.NewZoneAddress(HostChainId)
	zoneAcc := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, zoneAddress)
	s.App.AccountKeeper.SetAccount(s.Ctx, zoneAcc)
	s.FundAccount(zoneAddress, sdk.NewInt64Coin(StAtom, 1_000_000))
	s.FundAccount(addr0, sdk.NewInt64Coin(IbcAtom, 10_000_000))

	// Set up host zone
	hostZone := types.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Address:        zoneAddress.String(),
	}
	epochTracker := types.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	// Set up deposit record for MsgLiquidStake
	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             sdkmath.NewInt(1_000_000),
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord)

	// Set up unbonding record for MsgRedeemStake
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}

	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.NewInt(1000000),
		Denom:             Atom,
		HostZoneId:        HostChainId,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	return priv0, addr0
}

// setUpBenchmarkSuite sets up a app test suite, tx, and post handler for benchmark tests.
// It returns the app configured to the correct state, a valid tx, and the protorev post handler.
func setUpBenchmarkSuite(msgs []sdk.Msg, privKey cryptotypes.PrivKey) (*KeeperTestSuite, authsigning.Tx, sdk.AnteHandler, sdk.PostHandler) {
	// Create a new test suite
	s := new(KeeperTestSuite)
	s.SetT(&testing.T{})
	priv0, _ := s.SetUpHostZoneAndAccount(privKey)

	// Build the tx
	// _, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}

	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	tx := s.BuildTx(txBuilder, msgs, "", sdk.NewCoins(sdk.NewCoin("ustrd", sdk.NewInt(10000))), 500000, priv0, s.clientCtx.TxConfig)

	// Set up the ante handler
	stakeIbcAnteDecorator := keeper.NewStakeIbcAnteDecorator(s.App.StakeibcKeeper)
	anteHandlerStakeIbc := sdk.ChainAnteDecorators(stakeIbcAnteDecorator)
	// Set up the post handler
	stakeIbcDecorator := keeper.NewStakeIbcPostDecorator(s.App.StakeibcKeeper)
	posthandlerStakeIbc := sdk.ChainPostDecorators(stakeIbcDecorator)

	return s, tx, anteHandlerStakeIbc, posthandlerStakeIbc
}
