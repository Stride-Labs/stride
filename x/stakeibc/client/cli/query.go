package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
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
	cmd.AddCommand(CmdShowValidators())
	cmd.AddCommand(CmdListHostZone())
	cmd.AddCommand(CmdShowHostZone())
	cmd.AddCommand(CmdModuleAddress())
	cmd.AddCommand(CmdShowInterchainAccount())
	cmd.AddCommand(CmdListEpochTracker())
	cmd.AddCommand(CmdShowEpochTracker())
	cmd.AddCommand(CmdNextPacketSequence())

	return cmd
}
