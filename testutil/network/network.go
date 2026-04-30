package network

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	cometbftrand "github.com/cometbft/cometbft/libs/rand"
	cosmosdb "github.com/cosmos/cosmos-db"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v32/app"
)

type (
	Network   = network.Network
	Config    = network.Config
	Validator = network.Validator
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	t.Helper()
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, _ := network.New(t, t.TempDir(), cfg)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	tempApp := app.InitStrideTestApp(false)
	encoding := app.MakeEncodingConfig()
	bondedTokens := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)

	chainId := fmt.Sprintf("stride-%d", cometbftrand.NewRand().Uint64())
	genState := tempApp.DefaultGenesis()
	return network.Config{
		Codec:             encoding.Codec,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			// Wire the network's single validator into POA + staking + bank.
			// Post-v33 (ICS → POA migration) this is what makes the chain
			// produce blocks: POA is the InitChain validator-update source,
			// and a Bonded staking shadow record is what distribution looks
			// up at every BeginBlock past height 1. See seedNetworkValidator
			// for the full rationale.
			if err := seedNetworkValidator(val.(network.Validator), bondedTokens); err != nil {
				panic(err)
			}

			// Drop gentxs so the chain's validator set comes only from POA's
			// genesis, not from a framework-generated staking gentx.
			if err := clearGentxs(val.(network.Validator)); err != nil {
				panic(err)
			}

			// A custom home directory is needed for wasm tests since wasmvm locks the directory
			tempHomeDir, err := os.MkdirTemp("", "stride-unit-test")
			if err != nil {
				panic(err)
			}

			return app.NewStrideApp(
				val.GetCtx().Logger,
				cosmosdb.NewMemDB(),
				true,
				simtestutil.NewAppOptionsWithFlagHome(tempHomeDir),
				[]wasmkeeper.Option{},
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID(chainId),
			)
		},
		GenesisState:    genState,
		TimeoutCommit:   2 * time.Second,
		ChainID:         chainId,
		NumValidators:   1,
		BondDenom:       sdk.DefaultBondDenom,
		MinGasPrices:    fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    bondedTokens,
		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}
}

// seedNetworkValidator wires the network's single validator into every module
// that needs to know about it for blocks to flow. There are two demands:
//
//   - POA must hold the validator so its InitGenesis returns the InitChain
//     ValidatorUpdate. Post-v33 (ICS → POA migration) ccvconsumer is no longer
//     in the module manager, so POA is the only source — without this seed
//     CometBFT panics on an empty validator set.
//
//   - Staking must hold a Bonded shadow record at the same consensus key, with
//     the bonded pool funded to match. distribution.AllocateTokens calls
//     stakingKeeper.ValidatorByConsAddr at every BeginBlock past height 1, and
//     post-v33 POA doesn't shadow validators into staking — so without this
//     record the chain consensus-fails at block 3 with "validator does not
//     exist". Pre-v33 ccvconsumer's hooks created the shadow automatically.
//
// One function, three module sections (poa, staking, bank), one round-trip
// through the genesis file.
func seedNetworkValidator(val network.Validator, bondAmt sdkmath.Int) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis from the file")
	}

	pkAny, err := codectypes.NewAnyWithValue(val.PubKey)
	if err != nil {
		return errorsmod.Wrap(err, "failed to wrap validator pubkey as Any")
	}

	// --- POA: validator-set source for InitChain ---
	poaGenesis := &poatypes.GenesisState{
		Params: poatypes.Params{Admin: authtypes.NewModuleAddress("gov").String()},
		Validators: []poatypes.Validator{
			{
				PubKey: pkAny,
				Power:  bondAmt.Quo(sdk.DefaultPowerReduction).Int64(),
				Metadata: &poatypes.ValidatorMetadata{
					OperatorAddress: val.Address.String(),
					Moniker:         "test-validator",
				},
			},
		},
	}
	poaBz, err := val.ClientCtx.Codec.MarshalJSON(poaGenesis)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal POA genesis state into JSON")
	}
	appState[poatypes.ModuleName] = poaBz

	// --- Staking: Bonded shadow record so distribution can resolve the proposer ---
	// Mark this as exported and leave LastValidatorPowers empty so staking seeds
	// its indexes/state without returning InitGenesis validator updates. POA must
	// remain the sole validator-update source at genesis.
	stakingGenesis := stakingtypes.NewGenesisState(
		stakingtypes.DefaultParams(),
		[]stakingtypes.Validator{{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{Moniker: "test-validator"},
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}},
		[]stakingtypes.Delegation{
			stakingtypes.NewDelegation(
				val.Address.String(),
				sdk.ValAddress(val.Address).String(),
				sdkmath.LegacyOneDec(),
			),
		},
	)
	stakingGenesis.LastTotalPower = bondAmt
	stakingGenesis.Exported = true

	stakingBz, err := val.ClientCtx.Codec.MarshalJSON(stakingGenesis)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal staking genesis state into JSON")
	}
	appState[stakingtypes.ModuleName] = stakingBz

	// --- Bank: move the validator's self-bond into the bonded pool, then recompute supply ---
	var bankGenesis banktypes.GenesisState
	if err := val.ClientCtx.Codec.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return errorsmod.Wrap(err, "failed to unmarshal bank genesis")
	}

	if err := moveBondedTokens(&bankGenesis, val.Address.String(), bondAmt); err != nil {
		return err
	}

	totalSupply := sdk.NewCoins()
	for _, b := range bankGenesis.Balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}
	bankGenesis.Supply = totalSupply
	bankBz, err := val.ClientCtx.Codec.MarshalJSON(&bankGenesis)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal bank genesis state into JSON")
	}
	appState[banktypes.ModuleName] = bankBz

	return writeGenesis(genDoc, appState, genFile)
}

func moveBondedTokens(bankGenesis *banktypes.GenesisState, delegatorAddr string, bondAmt sdkmath.Int) error {
	bondDenom := sdk.DefaultBondDenom
	bondedPoolAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String()

	delegatorFound := false
	bondedPoolFound := false

	for i := range bankGenesis.Balances {
		balance := &bankGenesis.Balances[i]

		switch balance.Address {
		case delegatorAddr:
			currentBond := balance.Coins.AmountOf(bondDenom)
			if currentBond.LT(bondAmt) {
				return fmt.Errorf(
					"validator account %s has %s%s, need %s%s for bonded stake",
					delegatorAddr, currentBond.String(), bondDenom, bondAmt.String(), bondDenom,
				)
			}

			balance.Coins = balance.Coins.Sub(sdk.NewCoin(bondDenom, bondAmt))
			delegatorFound = true

		case bondedPoolAddr:
			balance.Coins = balance.Coins.Add(sdk.NewCoin(bondDenom, bondAmt))
			bondedPoolFound = true
		}
	}

	if !delegatorFound {
		return fmt.Errorf("validator account balance not found in bank genesis: %s", delegatorAddr)
	}

	if !bondedPoolFound {
		bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
			Address: bondedPoolAddr,
			Coins:   sdk.NewCoins(sdk.NewCoin(bondDenom, bondAmt)),
		})
	}

	return nil
}

// clearGentxs replaces the genutil genesis with the default (empty gentx list).
// Called so the chain's validator set is taken from POA's genesis only, not
// from a framework-generated staking gentx that would compete with POA.
func clearGentxs(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis from the file")
	}

	genutilBz, err := val.ClientCtx.Codec.MarshalJSON(genutiltypes.DefaultGenesisState())
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal genutil genesis state into JSON")
	}
	appState[genutiltypes.ModuleName] = genutilBz

	return writeGenesis(genDoc, appState, genFile)
}

// writeGenesis serializes appState back into genDoc and writes it to genFile.
func writeGenesis(genDoc *genutiltypes.AppGenesis, appState map[string]json.RawMessage, genFile string) error {
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal application genesis state into JSON")
	}
	genDoc.AppState = appStateJSON
	if err := genutil.ExportGenesisFile(genDoc, genFile); err != nil {
		return errorsmod.Wrap(err, "failed to export genesis state")
	}
	return nil
}
