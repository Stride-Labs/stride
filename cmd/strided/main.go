package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/Stride-Labs/stride/v3/cmd/strided/cmd"
	"github.com/Stride-Labs/stride/v3/cmd/strided/cmd/config"

	"github.com/Stride-Labs/stride/v3/app"
)

func main() {
	config.SetupConfig()
	config.RegisterDenoms()

	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
