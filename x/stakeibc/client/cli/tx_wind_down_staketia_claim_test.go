// x/stakeibc/client/cli/tx_wind_down_staketia_claim_test.go
package cli_test

import (
	"testing"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/client/cli"
)

func TestCmdTransferStaketiaClaimBalance(t *testing.T) {
	ExecuteCLIExpectError(t, cli.CmdTransferStaketiaClaimBalance(), []string{"banana"}, "can not convert string to int")
}
