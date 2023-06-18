package main

import (
	"os"

	"github.com/Stride-Labs/stride/v10/app"
	cmdcfg "github.com/Stride-Labs/stride/v10/cmd/strided/config"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	cmdcfg.SetupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
