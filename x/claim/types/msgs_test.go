package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v33/x/claim/types"
)

func validCreateAirdropMsg() types.MsgCreateAirdrop {
	return types.MsgCreateAirdrop{
		Distributor: sdk.AccAddress([]byte("distributor_address_")).String(),
		Identifier:  "stride",
		ChainId:     "stride-1",
		Denom:       sdk.DefaultBondDenom,
		StartTime:   1,
		Duration:    1,
	}
}

func TestMsgCreateAirdrop_ValidateBasic(t *testing.T) {
	validMsg := validCreateAirdropMsg()
	require.NoError(t, validMsg.ValidateBasic(), "valid denom should pass")

	// An empty denom is rejected
	emptyDenom := validCreateAirdropMsg()
	emptyDenom.Denom = ""
	require.ErrorContains(t, emptyDenom.ValidateBasic(), "denom not set")

	// An SDK-invalid denom is rejected up front, so it can't later panic the bank
	// keeper's GetBalance via NewCoin in the claim hooks
	invalidDenom := validCreateAirdropMsg()
	invalidDenom.Denom = "!"
	require.ErrorContains(t, invalidDenom.ValidateBasic(), "invalid airdrop denom")
}
