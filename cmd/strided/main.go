package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/Stride-Labs/stride/v4/app"

	cmdcfg "github.com/Stride-Labs/stride/v4/cmd/strided/config"
)

func main() {
	cmdcfg.SetupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
