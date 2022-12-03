package cli_test

import (
	"fmt"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/testutil/network"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/client/cli"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func networkWithEpochTracker(t *testing.T) *network.Network {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg)
}

func TestShowEpochTracker(t *testing.T) {
	net := networkWithEpochTracker(t)
	ctx := net.Validators[0].ClientCtx
	strideEpochId := "stride_epoch"
	nonExistentId := "nonexistent_id"
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc              string
		idEpochIdentifier string

		args []string
		err  error
		obj  types.EpochTracker
	}{
		{
			desc:              "found",
			idEpochIdentifier: strideEpochId,

			args: common,
			obj:  types.EpochTracker{EpochIdentifier: strideEpochId, EpochNumber: 1},
		},
		{
			desc:              "not found",
			idEpochIdentifier: nonExistentId,

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idEpochIdentifier,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowEpochTracker(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetEpochTrackerResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.EpochTracker)
				require.Equal(t, tc.obj.EpochIdentifier, resp.EpochTracker.EpochIdentifier)
				require.Equal(t, tc.obj.EpochNumber, resp.EpochTracker.EpochNumber)

			}
		})
	}
}

func TestListEpochTracker(t *testing.T) {
	net := networkWithEpochTracker(t)
	ctx := net.Validators[0].ClientCtx
	out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListEpochTracker(), []string{})

	expected := []types.EpochTracker{
		{EpochIdentifier: "day", EpochNumber: 1},
		{EpochIdentifier: "mint", EpochNumber: 1},
		{EpochIdentifier: "stride_epoch", EpochNumber: 1},
		{EpochIdentifier: "week", EpochNumber: 1},
	}
	require.NoError(t, err)

	var actual types.QueryAllEpochTrackerResponse
	require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &actual))

	require.NotNil(t, actual.EpochTracker)
	require.Len(t, actual.EpochTracker, 4)

	actualTrim := []types.EpochTracker{}
	for _, epochTracker := range actual.EpochTracker {
		trimmed := types.EpochTracker{
			EpochIdentifier: epochTracker.EpochIdentifier,
			EpochNumber:     epochTracker.EpochNumber,
		}
		actualTrim = append(actualTrim, trimmed)
	}
	require.Equal(t, expected, actualTrim)
}
