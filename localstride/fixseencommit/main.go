// Command fixseencommit repairs the seen commit that `strided in-place-testnet` would
// otherwise corrupt. It is a prerequisite of the localstride upgrade flow - see
// localstride/README.md.
//
// The cosmos-sdk testnetify() routine rewrites the seen commit at the node's last block so
// that it appears signed by the local validator (server/start.go):
//
//	seenCommit.Signatures[0].Signature        = vote.Signature
//	seenCommit.Signatures[0].ValidatorAddress = validatorAddress
//	seenCommit.Signatures[0].Timestamp        = vote.Timestamp
//
// It assigns the validator address but never touches BlockIDFlag, assuming slot 0 belongs to
// a validator that signed the block. When that validator was absent, the flag stays
// BlockIDFlagAbsent while the address becomes non-empty - precisely the combination
// CommitSig.ValidateBasic() rejects with "validator address is present". The node then panics
// on the next start, while reconstructing its last commit.
//
// Promoting a validator that actually signed into slot 0 beforehand avoids this: testnetify
// overwrites every other field of that signature, so only the COMMIT flag needs to survive.
// Running this against an already-healthy seen commit is a no-op.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	cmtcfg "github.com/cometbft/cometbft/config"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
)

func main() {
	defaultHome := filepath.Join(os.Getenv("HOME"), ".stride-localstride")

	home := flag.String("home", defaultHome, "node home directory")
	height := flag.Int64("height", 0, "height to repair (defaults to the block store height)")
	dryRun := flag.Bool("dry-run", false, "report what would change without writing")
	flag.Parse()

	if err := repairSeenCommit(*home, *height, *dryRun); err != nil {
		log.Fatalf("failed to repair seen commit: %v", err)
	}
}

func repairSeenCommit(home string, height int64, dryRun bool) error {
	config, err := loadConfig(home)
	if err != nil {
		return err
	}

	blockStoreDB, err := cmtcfg.DefaultDBProvider(&cmtcfg.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return fmt.Errorf("opening block store: %w", err)
	}
	defer blockStoreDB.Close()

	// Match testnetify(), which keys off state.LastBlockHeight. That trails the block store
	// height whenever the node stopped mid-block - as it does when halting for an upgrade.
	if height == 0 {
		height, err = lastBlockHeight(config)
		if err != nil {
			return err
		}
	}

	key := []byte(fmt.Sprintf("SC:%v", height))
	bz, err := blockStoreDB.Get(key)
	if err != nil {
		return fmt.Errorf("reading seen commit at height %d: %w", height, err)
	}
	if len(bz) == 0 {
		return fmt.Errorf("no seen commit stored at height %d", height)
	}

	commit := new(cmtproto.Commit)
	if err := proto.Unmarshal(bz, commit); err != nil {
		return fmt.Errorf("unmarshalling seen commit at height %d: %w", height, err)
	}
	if len(commit.Signatures) == 0 {
		return fmt.Errorf("seen commit at height %d has no signatures", height)
	}

	if commit.Signatures[0].BlockIdFlag == cmtproto.BlockIDFlagCommit {
		fmt.Printf("Seen commit at height %d is already signed in slot 0 - nothing to do\n", height)
		return nil
	}

	signer, err := findSigner(commit)
	if err != nil {
		return fmt.Errorf("at height %d: %w", height, err)
	}

	fmt.Printf("Seen commit at height %d has %s in slot 0\n", height, commit.Signatures[0].BlockIdFlag)
	fmt.Printf("Promoting signature %d (validator %X) into slot 0\n", signer, commit.Signatures[signer].ValidatorAddress)

	if dryRun {
		fmt.Println("Dry run - no changes written")
		return nil
	}

	commit.Signatures[0], commit.Signatures[signer] = commit.Signatures[signer], commit.Signatures[0]

	repaired, err := proto.Marshal(commit)
	if err != nil {
		return fmt.Errorf("marshalling repaired seen commit: %w", err)
	}
	if err := blockStoreDB.SetSync(key, repaired); err != nil {
		return fmt.Errorf("writing repaired seen commit: %w", err)
	}

	fmt.Println("Done")
	return nil
}

// lastBlockHeight reads state.LastBlockHeight straight out of the state store, rather than via
// the state package, so that a store written by an older CometBFT still loads.
func lastBlockHeight(config *cmtcfg.Config) (int64, error) {
	stateDB, err := cmtcfg.DefaultDBProvider(&cmtcfg.DBContext{ID: "state", Config: config})
	if err != nil {
		return 0, fmt.Errorf("opening state store: %w", err)
	}
	defer stateDB.Close()

	bz, err := stateDB.Get([]byte("stateKey"))
	if err != nil {
		return 0, fmt.Errorf("reading state: %w", err)
	}
	if len(bz) == 0 {
		return 0, fmt.Errorf("no state stored in %s", config.DBDir())
	}

	state := new(cmtstate.State)
	if err := proto.Unmarshal(bz, state); err != nil {
		return 0, fmt.Errorf("unmarshalling state: %w", err)
	}

	return state.LastBlockHeight, nil
}

// loadConfig reads the node's config.toml so the block store is opened with the configured
// backend, rather than assuming the CometBFT default.
func loadConfig(home string) (*cmtcfg.Config, error) {
	reader := viper.New()
	reader.SetConfigFile(filepath.Join(home, "config", "config.toml"))
	if err := reader.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config.toml: %w", err)
	}

	config := cmtcfg.DefaultConfig()
	if err := reader.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("parsing config.toml: %w", err)
	}
	config.SetRoot(home)

	return config, nil
}

// findSigner returns the index of a validator that voted for the block, and so carries a
// signature and timestamp that survive CommitSig validation once moved into slot 0.
func findSigner(commit *cmtproto.Commit) (int, error) {
	for i, signature := range commit.Signatures {
		if signature.BlockIdFlag == cmtproto.BlockIDFlagCommit {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no validator signed the block, cannot repair")
}
