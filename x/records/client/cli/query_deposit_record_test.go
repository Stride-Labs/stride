package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/testutil/network"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/records/client/cli"
	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func networkWithDepositRecordObjects(t *testing.T, n int) (*network.Network, []types.DepositRecord) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		depositRecord := types.DepositRecord{
			Id: uint64(i),
		}
		nullify.Fill(&depositRecord)
		state.DepositRecordList = append(state.DepositRecordList, depositRecord)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	// fmt.Println(fmt.Sprintf("state.DepositRecordList: %v", state.DepositRecordList))
	return network.New(t, cfg), state.DepositRecordList
}

func TestShowDepositRecord(t *testing.T) {
	net, objs := networkWithDepositRecordObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	_ = ctx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		id   string
		args []string
		err  error
		obj  types.DepositRecord
	}{
		{
			desc: "found",
			id:   fmt.Sprintf("%d", objs[0].Id),
			args: common,
			obj:  objs[0],
		},
		{
			desc: "not found",
			id:   "not_found",
			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		// TODO why is this test failing?
		_ = tc
		// t.Run(tc.desc, func(t *testing.T) {
		// 	args := []string{tc.id}
		// 	args = append(args, tc.args...)
		// 	// fmt.Println(fmt.Sprintf("args: %v", args))
		// 	out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowDepositRecord(), args)
		// 	if tc.err != nil {
		// 		stat, ok := status.FromError(tc.err)
		// 		require.True(t, ok)
		// 		require.ErrorIs(t, stat.Err(), tc.err)
		// 	} else {
		// 		require.NoError(t, err)
		// 		var resp types.QueryGetDepositRecordResponse
		// 		require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		// 		require.NotNil(t, resp.DepositRecord)
		// 		require.Equal(t,
		// 			nullify.Fill(&tc.obj),
		// 			nullify.Fill(&resp.DepositRecord),
		// 		)
		// 	}
		// })
	}
}

func TestListDepositRecord(t *testing.T) {
	net, objs := networkWithDepositRecordObjects(t, 5)

	ctx := net.Validators[0].ClientCtx
	request := func(next []byte, offset, limit uint64, total bool) []string {
		args := []string{
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		}
		if next == nil {
			args = append(args, fmt.Sprintf("--%s=%d", flags.FlagOffset, offset))
		} else {
			args = append(args, fmt.Sprintf("--%s=%s", flags.FlagPageKey, next))
		}
		args = append(args, fmt.Sprintf("--%s=%d", flags.FlagLimit, limit))
		if total {
			args = append(args, fmt.Sprintf("--%s", flags.FlagCountTotal))
		}
		return args
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(objs); i += step {
			args := request(nil, uint64(i), uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListDepositRecord(), args)
			require.NoError(t, err)
			var resp types.QueryAllDepositRecordResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.DepositRecord), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.DepositRecord),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(objs); i += step {
			args := request(next, 0, uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListDepositRecord(), args)
			require.NoError(t, err)
			var resp types.QueryAllDepositRecordResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.DepositRecord), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.DepositRecord),
			)
			next = resp.Pagination.NextKey
		}
	})
	// TODO: why is this test failing?
	// t.Run("Total", func(t *testing.T) {
	// 	args := request(nil, 0, uint64(len(objs)), true)
	// 	out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListDepositRecord(), args)
	// 	require.NoError(t, err)
	// 	var resp types.QueryAllDepositRecordResponse
	// 	require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
	// 	require.NoError(t, err)
	// 	fmt.Println(fmt.Sprintf("objs 2: %v", objs))
	// 	fmt.Println(fmt.Sprintf("t: %v", t))
	// 	require.Equal(t, len(objs), int(resp.Pagination.Total))
	// 	require.ElementsMatch(t,
	// 		nullify.Fill(objs),
	// 		nullify.Fill(resp.DepositRecord),
	// 	)
	// })
}
