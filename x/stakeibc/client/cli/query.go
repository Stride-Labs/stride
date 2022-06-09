package cli

import (
	"fmt"
	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group stakeibc queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdShowValidator())
	cmd.AddCommand(CmdShowDelegation())
	cmd.AddCommand(CmdShowMinValidatorRequirements())
	cmd.AddCommand(CmdShowICAAccount())
	cmd.AddCommand(CmdListHostZone())
	cmd.AddCommand(CmdShowHostZone())
	// this line is used by starport scaffolding # 1

	return cmd
}
