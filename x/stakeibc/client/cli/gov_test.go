package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v30/x/stakeibc/client/cli"
)

func TestCmdAddValidatorsProposal(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		args := []string{
			"[proposal-file]",
		}

		cmd := cli.CmdAddValidatorsProposal()
		ExecuteCLIExpectError(t, cmd, args, `open [proposal-file]: no such file or directory`)
	})
	t.Run("empty file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()

		args := []string{
			f.Name(),
		}

		cmd := cli.CmdAddValidatorsProposal()
		ExecuteCLIExpectError(t, cmd, args, `EOF`)
	})
	t.Run("non json file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString("This is not JSON")
		require.NoError(t, err)

		args := []string{
			f.Name(),
		}

		cmd := cli.CmdAddValidatorsProposal()
		ExecuteCLIExpectError(t, cmd, args, `invalid character 'T' looking for beginning of value`)
	})
	t.Run("wrong json format", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(`{"description":"Proposal to add Imperator because they contribute in XYZ ways!","hostZone":"GAIA","blabla_validators":[{"name":"Imperator","address":"cosmosvaloper1v5y0tg0jllvxf5c3afml8s3awue0ymju89frut"}],"deposit":"64000000ustrd"}`)
		require.NoError(t, err)

		args := []string{
			f.Name(),
		}

		cmd := cli.CmdAddValidatorsProposal()
		ExecuteCLIExpectError(t, cmd, args, `unknown field "blabla_validators" in types.AddValidatorsProposal`)
	})
}

func TestCmdToggleLSMProposal(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		args := []string{
			"[proposal-file]",
		}

		cmd := cli.CmdToggleLSMProposal()
		ExecuteCLIExpectError(t, cmd, args, `open [proposal-file]: no such file or directory`)
	})
	t.Run("empty file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()

		args := []string{
			f.Name(),
		}

		cmd := cli.CmdToggleLSMProposal()
		ExecuteCLIExpectError(t, cmd, args, `EOF`)
	})
	t.Run("non json file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString("This is not JSON")
		require.NoError(t, err)

		args := []string{
			f.Name(),
		}

		cmd := cli.CmdToggleLSMProposal()
		ExecuteCLIExpectError(t, cmd, args, `invalid character 'T' looking for beginning of value`)
	})
	t.Run("wrong json format", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(`{"description":"Proposal to add Imperator because they contribute in XYZ ways!","hostZone":"GAIA","blabla_validators":[{"name":"Imperator","address":"cosmosvaloper1v5y0tg0jllvxf5c3afml8s3awue0ymju89frut"}],"deposit":"64000000ustrd"}`)
		require.NoError(t, err)

		args := []string{
			f.Name(),
		}

		cmd := cli.CmdToggleLSMProposal()
		ExecuteCLIExpectError(t, cmd, args, `unknown field "blabla_validators" in types.ToggleLSMProposal`)
	})
}
