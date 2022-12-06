package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	claimTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	claimTxCmd.AddCommand(CmdClaimFreeAmount())
	claimTxCmd.AddCommand(CmdSetAirdropAllocations())
	claimTxCmd.AddCommand(CmdCreateAirdrop())
	claimTxCmd.AddCommand(CmdDeleteAirdrop())
	return claimTxCmd
}
