package cli

import (
	"github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

// GetQueryCmd creates and returns the interchainquery query command
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the interchainquery module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// cmd.AddCommand(getInterchainAccountCmd())
	return cmd
}
