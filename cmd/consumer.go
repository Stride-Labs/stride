package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	types1 "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	pvm "github.com/cometbft/cometbft/privval"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/v6/x/ccv/consumer/types"
	ccvtypes "github.com/cosmos/interchain-security/v6/x/ccv/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v26/testutil"
)

const (
	FlagValidatorPublicKeys      = "validator-public-keys"
	FlagValidatorHomeDirectories = "validator-home-directories"
)

// Builds the list of validator Ed25519 pubkeys from a comma separate list of base64 encoded pubkeys
func buildPublicKeysFromString(publicKeysRaw string) (publicKeys []tmprotocrypto.PublicKey, err error) {
	for _, publicKeyEncoded := range strings.Split(publicKeysRaw, ",") {
		if publicKeyEncoded == "" {
			continue
		}
		publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyEncoded)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "unable to decode public key")
		}
		publicKeys = append(publicKeys, tmprotocrypto.PublicKey{
			Sum: &tmprotocrypto.PublicKey_Ed25519{
				Ed25519: publicKeyBytes,
			},
		})
	}

	return publicKeys, nil
}

// Builds the list validator Ed25519 pubkeys from a comma separated list of validator home directories
func buildPublicKeysFromHomeDirectories(config *config.Config, homeDirectories string) (publicKeys []tmprotocrypto.PublicKey, err error) {
	for _, homeDir := range strings.Split(homeDirectories, ",") {
		if homeDir == "" {
			continue
		}
		config.SetRoot(homeDir)

		privValidator := pvm.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
		pk, err := privValidator.GetPubKey()
		if err != nil {
			return nil, err
		}
		sdkPublicKey, err := cryptocodec.FromTmPubKeyInterface(pk)
		if err != nil {
			return nil, err
		}
		tmProtoPublicKey, err := cryptocodec.ToTmProtoPublicKey(sdkPublicKey)
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, tmProtoPublicKey)
	}

	return publicKeys, nil
}

func AddConsumerSectionCmd(defaultNodeHome string) *cobra.Command {
	genesisMutator := NewDefaultGenesisIO()

	cmd := &cobra.Command{
		Use:                        "add-consumer-section",
		Args:                       cobra.ExactArgs(0),
		Short:                      "ONLY FOR TESTING PURPOSES! Modifies genesis so that chain can be started locally with one node.",
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			// We need to public keys for each validator - they can either be provided explicitly
			// or indirectly by providing the validator home directories
			publicKeysRaw, err := cmd.Flags().GetString(FlagValidatorPublicKeys)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse validator public key flag")
			}
			homeDirectoriesRaw, err := cmd.Flags().GetString(FlagValidatorHomeDirectories)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse validator home directories flag")
			}
			if (publicKeysRaw == "" && homeDirectoriesRaw == "") || (publicKeysRaw != "" && homeDirectoriesRaw != "") {
				return fmt.Errorf("must specified either --%s or --%s", FlagValidatorPublicKeys, FlagValidatorHomeDirectories)
			}

			// Build up a list of the validator public keys
			// If the public keys were passed directly, decode them and create the tm proto pub keys
			// Otherwise, derrive them from the private keys in each validator's home directory
			var tmPublicKeys []tmprotocrypto.PublicKey
			if publicKeysRaw != "" {
				tmPublicKeys, err = buildPublicKeysFromString(publicKeysRaw)
				if err != nil {
					return err
				}
			} else {
				serverCtx := server.GetServerContextFromCmd(cmd)
				config := serverCtx.Config

				tmPublicKeys, err = buildPublicKeysFromHomeDirectories(config, homeDirectoriesRaw)
				if err != nil {
					return err
				}
			}

			if len(tmPublicKeys) == 0 {
				return errors.New("no valid public keys or validator home directories provided")
			}

			return genesisMutator.AlterConsumerModuleState(cmd, func(state *GenesisData, _ map[string]json.RawMessage) error {
				initialValset := []types1.ValidatorUpdate{}
				genesisState := testutil.CreateMinimalConsumerTestGenesis()

				for _, publicKey := range tmPublicKeys {
					initialValset = append(initialValset, types1.ValidatorUpdate{PubKey: publicKey, Power: 100})
				}

				vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValset)
				if err != nil {
					return errorsmod.Wrap(err, "could not convert val updates to validator set")
				}

				genesisState.Provider.InitialValSet = initialValset
				genesisState.Provider.ConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).Hash()

				state.ConsumerModuleState = genesisState
				return nil
			})
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(FlagValidatorPublicKeys, "", "Comma separated, base64-encoded public keys for each validator")
	cmd.Flags().String(FlagValidatorHomeDirectories, "", "Comma separated list of home directories for each validator")

	return cmd
}

type GenesisMutator interface {
	AlterConsumerModuleState(cmd *cobra.Command, callback func(state *GenesisData, appState map[string]json.RawMessage) error) error
}

type DefaultGenesisIO struct {
	DefaultGenesisReader
}

func NewDefaultGenesisIO() *DefaultGenesisIO {
	return &DefaultGenesisIO{DefaultGenesisReader: DefaultGenesisReader{}}
}

func (x DefaultGenesisIO) AlterConsumerModuleState(cmd *cobra.Command, callback func(state *GenesisData, appState map[string]json.RawMessage) error) error {
	g, err := x.ReadGenesis(cmd)
	if err != nil {
		return err
	}
	if err := callback(g, g.AppState); err != nil {
		return err
	}
	if err := g.ConsumerModuleState.Validate(); err != nil {
		return err
	}
	clientCtx := client.GetClientContextFromCmd(cmd)
	consumerGenStateBz, err := clientCtx.Codec.MarshalJSON(g.ConsumerModuleState)
	if err != nil {
		return errorsmod.Wrap(err, "marshal consumer genesis state")
	}

	g.AppState[ccvconsumertypes.ModuleName] = consumerGenStateBz
	appStateJSON, err := json.Marshal(g.AppState)
	if err != nil {
		return errorsmod.Wrap(err, "marshal application genesis state")
	}

	g.GenDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(g.GenDoc, g.GenesisFile)
}

type DefaultGenesisReader struct{}

func (d DefaultGenesisReader) ReadGenesis(cmd *cobra.Command) (*GenesisData, error) {
	clientCtx := client.GetClientContextFromCmd(cmd)
	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config
	config.SetRoot(clientCtx.HomeDir)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	return NewGenesisData(
		genFile,
		genDoc,
		appState,
		nil,
	), nil
}

type GenesisData struct {
	GenesisFile         string
	GenDoc              *tmtypes.GenesisDoc
	AppState            map[string]json.RawMessage
	ConsumerModuleState *ccvtypes.ConsumerGenesisState
}

func NewGenesisData(genesisFile string, genDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage, consumerModuleState *ccvtypes.ConsumerGenesisState) *GenesisData {
	return &GenesisData{GenesisFile: genesisFile, GenDoc: genDoc, AppState: appState, ConsumerModuleState: consumerModuleState}
}
