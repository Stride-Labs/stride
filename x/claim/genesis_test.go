package claim_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

func TestGenesis(t *testing.T) {
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	pub2 := secp256k1.GenPrivKey().PubKey()
	addr2 := sdk.AccAddress(pub2.Address())

	pub3 := secp256k1.GenPrivKey().PubKey()
	addr3 := sdk.AccAddress(pub3.Address())

	genesisState := types.GenesisState{
		Params: types.Params{
			Airdrops: []*types.Airdrop{
				{
					AirdropIdentifier:  types.DefaultAirdropIdentifier,
					AirdropStartTime:   time.Now(),
					AirdropDuration:    types.DefaultAirdropDuration,
					ClaimDenom:         sdk.DefaultBondDenom,
					DistributorAddress: addr3.String(),
				},
			},
		},
		ClaimRecords: []types.ClaimRecord{
			{
				Address:           addr1.String(),
				Weight:            sdk.NewDecWithPrec(50, 2), // 50%
				ActionCompleted:   []bool{false, false, false},
				AirdropIdentifier: types.DefaultAirdropIdentifier,
			},
			{
				Address:           addr2.String(),
				Weight:            sdk.NewDecWithPrec(50, 2), // 50%
				ActionCompleted:   []bool{false, false, false},
				AirdropIdentifier: "juno",
			},
		},
	}

	k, ctx := keepertest.ClaimKeeper(t)
	k.InitGenesis(ctx, genesisState)
	got := k.ExportGenesis(ctx)
	require.NotNil(t, got)

	totalWeightStride, err := k.GetTotalWeight(ctx, types.DefaultAirdropIdentifier)
	require.NoError(t, err)
	require.Equal(t, totalWeightStride, genesisState.ClaimRecords[0].Weight)

	totalWeightJuno, err := k.GetTotalWeight(ctx, types.DefaultAirdropIdentifier)
	require.NoError(t, err)
	require.Equal(t, totalWeightJuno, genesisState.ClaimRecords[1].Weight)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.Params, got.Params)
	require.Equal(t, genesisState.ClaimRecords, got.ClaimRecords)
}
