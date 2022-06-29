package cli_test

import (
	"fmt"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/testutil/network"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/stakeibc/client/cli"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func networkWithValidatorObjects(t *testing.T) (*network.Network, map[string][]*types.Validator) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	var validatorsByHostZone = make(map[string][]*types.Validator)
	validators := []*types.Validator{}
	nullify.Fill(&validators)

	chainId := "GAIA"
	hostZone := &types.HostZone{
		ChainId:    chainId,
		Validators: validators,
	}
	validatorsByHostZone[chainId] = validators

	state.HostZoneList = []types.HostZone{*hostZone}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), validatorsByHostZone
}

func TestShowValidator(t *testing.T) {
	net, validatorsByHostZone := networkWithValidatorObjects(t)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  []*types.Validator
	}{
		{
			desc: "get",
			args: append(common, "GAIA"),
			obj:  validatorsByHostZone["GAIA"],
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			var args []string
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowValidators(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetValidatorsResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Validators)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Validators),
				)
			}
		})
	}
}
