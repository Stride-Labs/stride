// x/stakeibc/client/cli/tx_wind_down_undelegate_test.go
package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/client/cli"
)

func TestCmdUndelegateFromValidators(t *testing.T) {
	t.Run("neither file nor --all", func(t *testing.T) {
		ExecuteCLIExpectError(t, cli.CmdUndelegateFromValidators(), []string{"cosmoshub-4"}, "pass a validators file or --all")
	})

	t.Run("both file and --all", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "validators.json")
		require.NoError(t, os.WriteFile(path, []byte(`[{"address":"cosmosvaloper1abc"}]`), 0o600))
		ExecuteCLIExpectError(t, cli.CmdUndelegateFromValidators(), []string{"cosmoshub-4", path, "--all"}, "not both")
	})

	t.Run("missing file", func(t *testing.T) {
		ExecuteCLIExpectError(t, cli.CmdUndelegateFromValidators(), []string{"cosmoshub-4", "/does/not/exist.json"}, "no such file")
	})

	t.Run("bad offset in file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "validators.json")
		require.NoError(t, os.WriteFile(path, []byte(`[{"address":"cosmosvaloper1abc","offset":"banana"}]`), 0o600))
		ExecuteCLIExpectError(t, cli.CmdUndelegateFromValidators(), []string{"cosmoshub-4", path}, "can not convert string to int")
	})
}
