// x/stakeibc/client/cli/tx_wind_down_transfer_from_ica_test.go
package cli_test

import (
	"testing"

	"github.com/Stride-Labs/stride/v35/x/stakeibc/client/cli"
)

func TestCmdTransferFromIca(t *testing.T) {
	t.Run("bad ica type", func(t *testing.T) {
		ExecuteCLIExpectError(t, cli.CmdTransferFromIca(), []string{"cosmoshub-4", "TREASURY", "1000uatom"}, "unknown ica type")
	})
	t.Run("bad amount", func(t *testing.T) {
		ExecuteCLIExpectError(t, cli.CmdTransferFromIca(), []string{"cosmoshub-4", "DELEGATION", "banana"}, "invalid decimal coin expression")
	})
}
