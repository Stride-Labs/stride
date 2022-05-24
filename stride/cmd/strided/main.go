package main

import (
	"os"

	"github.com/Stride-Labs/stride/app"
	svrcmd "github.com/Stride-Labs/cosmos-sdk/server/cmd"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"

	cmdcfg "github.com/Stride-Labs/stride/cmd/strided/config"
)

func main() {
	setupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

func setupConfig() {
	// set the address prefixes
	config := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(config)
	cmdcfg.SetBip44CoinType(config)
	config.Seal()
}
