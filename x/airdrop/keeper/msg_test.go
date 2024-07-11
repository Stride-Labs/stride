package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func TestMsgClaim(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	k = k // TODO remove this, just put it here to ignore the "k declared and not used" error

	// TODO init state
	// params := types.DefaultParams()
	// require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgClaim
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgClaim{
				Claimer: "invalid",
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			input: &types.MsgClaim{
				Claimer: "", // TODO get from initialized state
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.Claim(wctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			// TODO check state after msg execution
		})
	}
}

func TestMsgClaimAndStake(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	k = k // TODO remove this, just put it here to ignore the "k declared and not used" error

	// TODO init state
	// params := types.DefaultParams()
	// require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgClaimAndStake
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgClaimAndStake{
				Claimer: "invalid",
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			input: &types.MsgClaimAndStake{
				Claimer: "", // TODO get from initialized state
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.ClaimAndStake(wctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			// TODO check state after msg execution
		})
	}
}

func TestMsgClaimEarly(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	k = k // TODO remove this, just put it here to ignore the "k declared and not used" error

	// TODO init state
	// params := types.DefaultParams()
	// require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgClaimEarly
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgClaimEarly{
				Claimer: "invalid",
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			input: &types.MsgClaimEarly{
				Claimer: "", // TODO get from initialized state
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.ClaimEarly(wctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			// TODO check state after msg execution
		})
	}
}
