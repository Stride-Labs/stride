package cli_test

import (
	"fmt"
	"strconv"
	"testing"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v27/testutil/network"
	"github.com/Stride-Labs/stride/v27/testutil/nullify"
	"github.com/Stride-Labs/stride/v27/x/records/client/cli"
	"github.com/Stride-Labs/stride/v27/x/records/types"
)

// TODO [cleanup] - Migrate to new CLI testing framework
func networkWithUserRedemptionRecordObjects(t *testing.T, n int) (*network.Network, []types.UserRedemptionRecord) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		userRedemptionRecord := types.UserRedemptionRecord{
			Id:                strconv.Itoa(i),
			NativeTokenAmount: sdkmath.NewInt(int64(i)),
			StTokenAmount:     sdkmath.NewInt(int64(i)),
		}
		nullify.Fill(&userRedemptionRecord)
		state.UserRedemptionRecordList = append(state.UserRedemptionRecordList, userRedemptionRecord)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.UserRedemptionRecordList
}

func TestShowUserRedemptionRecord(t *testing.T) {
	net, objs := networkWithUserRedemptionRecordObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		id   string
		args []string
		err  error
		obj  types.UserRedemptionRecord
	}{
		{
			desc: "found",
			id:   objs[0].Id,
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
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.id}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowUserRedemptionRecord(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetUserRedemptionRecordResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.UserRedemptionRecord)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.UserRedemptionRecord),
				)
			}
		})
	}
}

func TestListUserRedemptionRecord(t *testing.T) {
	net, objs := networkWithUserRedemptionRecordObjects(t, 5)

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
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListUserRedemptionRecord(), args)
			require.NoError(t, err)
			var resp types.QueryAllUserRedemptionRecordResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.UserRedemptionRecord), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.UserRedemptionRecord),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(objs); i += step {
			args := request(next, 0, uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListUserRedemptionRecord(), args)
			require.NoError(t, err)
			var resp types.QueryAllUserRedemptionRecordResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.UserRedemptionRecord), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.UserRedemptionRecord),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		args := request(nil, 0, uint64(len(objs)), true)
		out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListUserRedemptionRecord(), args)
		require.NoError(t, err)
		var resp types.QueryAllUserRedemptionRecordResponse
		require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		require.NoError(t, err)
		require.Equal(t, len(objs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(objs),
			nullify.Fill(resp.UserRedemptionRecord),
		)
	})
}
