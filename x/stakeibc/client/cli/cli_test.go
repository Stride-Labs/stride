package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v27/testutil/network"
)

func ExecuteCLIExpectError(t *testing.T, cmd *cobra.Command, args []string, errorString string) {
	sdk.GetConfig().SetBech32PrefixForAccount("stride", "stridepub")

	clientCtx := client.Context{}.
		WithFromAddress(sdk.MustAccAddressFromBech32("stride10p3xzmnpdeshqctsv9ukzcm0vdhkuat52aucqd")).
		WithCodec(network.DefaultConfig().Codec)

	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.ErrorContains(t, err, errorString)
}
