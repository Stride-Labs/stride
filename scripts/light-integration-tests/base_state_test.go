package lighttest

import (
	"fmt"
	"testing"

	strideapp "github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestRegisterHostZone(t *testing.T) {
	t, coord, ctx := InitialBasicSetup(t)
	// goCtx := sdk.WrapSDKContext(ctx)

	if coord == nil {
		t.Error("coord is nil")
	}

	stride := coord.Chains[CHAIN_ID]
	app := (stride.App).(*strideapp.StrideApp)
	// k := app.StakeibcKeeper

	msg := &types.MsgRegisterHostZone{
		ConnectionId:       "connection-0",
		Bech32Prefix:       "cosmos",
		IbcDenom:           "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		Creator:            STRIDE_ACCT,
		TransferChannelId:  "transfer",
		UnbondingFrequency: 3,
	}
	handler := app.MsgServiceRouter().Handler(msg)
	resp, err := handler(ctx, msg)
	require.NoError(t, err)
	fmt.Printf("Response: %v\n", resp)
	// k.RegisterHostZone(goCtx)
}
