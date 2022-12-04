package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var DefaultRelativePacketTimeoutTimestamp = cast.ToUint64((time.Duration(10) * time.Minute).Nanoseconds())

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
	cmd.AddCommand(CmdRegisterHostZone())
	cmd.AddCommand(CmdRedeemStake())
	cmd.AddCommand(CmdClaimUndelegatedTokens())
	cmd.AddCommand(CmdRebalanceValidators())
	cmd.AddCommand(CmdAddValidator())
	cmd.AddCommand(CmdChangeValidatorWeight())
	cmd.AddCommand(CmdDeleteValidator())
	cmd.AddCommand(CmdRestoreInterchainAccount())
	cmd.AddCommand(CmdUpdateValidatorSharesExchRate())
	cmd.AddCommand(CmdClearBalance())
	// this line is used by starport scaffolding # 1

	return cmd
}
