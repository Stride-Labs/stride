package cli

import (
	"fmt"
	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/records/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group records queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListUserRedemptionRecord())
	cmd.AddCommand(CmdShowUserRedemptionRecord())
	cmd.AddCommand(CmdListEpochUnbondingRecord())
	cmd.AddCommand(CmdShowEpochUnbondingRecord())
	cmd.AddCommand(CmdListDepositRecord())
	cmd.AddCommand(CmdShowDepositRecord())
	cmd.AddCommand(CmdListDepositRecordByHost())
	cmd.AddCommand(CmdLSMDeposit())
	cmd.AddCommand(CmdLSMDeposits())

	return cmd
}
