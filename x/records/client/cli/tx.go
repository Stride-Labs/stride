package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/Stride-Labs/stride/v4/x/records/types"
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

	// this line is used by starport scaffolding # 1

	return cmd
}
