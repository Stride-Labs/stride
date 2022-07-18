package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdLiquidStake())
	// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
	cmd.AddCommand(CmdRegisterAccount())
	cmd.AddCommand(CmdSubmitTx())
	cmd.AddCommand(CmdRegisterHostZone())
	cmd.AddCommand(CmdRedeemStake())
	cmd.AddCommand(CmdClaimUndelegatedTokens())
	cmd.AddCommand(CmdRebalanceValidators())
	cmd.AddCommand(CmdAddValidator())
	cmd.AddCommand(CmdChangeValidatorWeight())
	cmd.AddCommand(CmdDeleteValidator())
	// this line is used by starport scaffolding # 1

	return cmd
}
