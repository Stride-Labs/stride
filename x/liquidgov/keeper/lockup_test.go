package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	config "github.com/Stride-Labs/stride/v5/cmd/strided/config"
	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
	pk2 = ed25519.GenPrivKey().PubKey()
	pk3 = ed25519.GenPrivKey().PubKey()

	addr1, _ = sdk.Bech32ifyAddressBytes(config.Bech32PrefixAccAddr, pk1.Address().Bytes())
	addr2, _ = sdk.Bech32ifyAddressBytes(config.Bech32PrefixAccAddr, pk2.Address().Bytes())
	addr3, _ = sdk.Bech32ifyAddressBytes(config.Bech32PrefixAccAddr, pk3.Address().Bytes())

	denom1 = "dnm1"
	denom2 = "dnm2"
	denom3 = "dnm3"
)

// tests GetLockup, GetCreatorLockups, SetLockup, RemoveLockup
func TestLockup(t *testing.T) {
	keeper, ctx := keepertest.LiquidgovKeeper(t)

	addrCreators := createIncrementalAccounts(3)

	// construct the denoms
	var denoms [3]string
	denoms[0] = denom1
	denoms[1] = denom2
	denoms[2] = denom3

	unlock1to1 := types.NewLockup(addrCreators[0], sdk.NewInt(9), denoms[0])

	// check the empty keeper first
	_, found := keeper.GetLockup(ctx, addrCreators[0], denoms[0])
	require.False(t, found)

	// set and retrieve a record
	keeper.SetLockup(ctx, unlock1to1)
	resBond, found := keeper.GetLockup(ctx, addrCreators[0], denoms[0])
	require.True(t, found)
	require.Equal(t, unlock1to1, resBond)

	// modify a records, save, and retrieve
	unlock1to1.Amount = sdk.NewInt(99)
	keeper.SetLockup(ctx, unlock1to1)
	resBond, found = keeper.GetLockup(ctx, addrCreators[0], denoms[0])
	require.True(t, found)
	require.Equal(t, unlock1to1, resBond)

	// add some more records
	unlock1to2 := types.NewLockup(addrCreators[0], sdk.NewInt(9), denoms[1])
	unlock1to3 := types.NewLockup(addrCreators[0], sdk.NewInt(9), denoms[2])
	unlock2to1 := types.NewLockup(addrCreators[1], sdk.NewInt(9), denoms[0])
	unlock2to2 := types.NewLockup(addrCreators[1], sdk.NewInt(9), denoms[1])
	unlock2to3 := types.NewLockup(addrCreators[1], sdk.NewInt(9), denoms[2])
	keeper.SetLockup(ctx, unlock1to2)
	keeper.SetLockup(ctx, unlock1to3)
	keeper.SetLockup(ctx, unlock2to1)
	keeper.SetLockup(ctx, unlock2to2)
	keeper.SetLockup(ctx, unlock2to3)

	// test all unlock retrieve capabilities
	resUnlocks := keeper.GetCreatorLockups(ctx, addrCreators[0], 5)
	require.Equal(t, 3, len(resUnlocks))
	require.Equal(t, unlock1to1, resUnlocks[0])
	require.Equal(t, unlock1to2, resUnlocks[1])
	require.Equal(t, unlock1to3, resUnlocks[2])
	resUnlocks = keeper.GetCreatorLockups(ctx, addrCreators[0], 2)
	require.Equal(t, 2, len(resUnlocks))
	resUnlocks = keeper.GetCreatorLockups(ctx, addrCreators[1], 5)
	require.Equal(t, 3, len(resUnlocks))
	require.Equal(t, unlock2to1, resUnlocks[0])
	require.Equal(t, unlock2to2, resUnlocks[1])
	require.Equal(t, unlock2to3, resUnlocks[2])
	allUnlocks := keeper.GetAllLockups(ctx)
	require.Equal(t, 6, len(allUnlocks))
	require.Equal(t, unlock1to1, allUnlocks[0])
	require.Equal(t, unlock1to2, allUnlocks[1])
	require.Equal(t, unlock1to3, allUnlocks[2])
	require.Equal(t, unlock2to1, allUnlocks[3])
	require.Equal(t, unlock2to2, allUnlocks[4])
	require.Equal(t, unlock2to3, allUnlocks[5])

	// delete a record
	keeper.RemoveLockup(ctx, unlock2to3)
	_, found = keeper.GetLockup(ctx, addrCreators[1], denoms[2])
	require.False(t, found)
	resUnlocks = keeper.GetCreatorLockups(ctx, addrCreators[1], 5)
	require.Equal(t, 2, len(resUnlocks))
	require.Equal(t, unlock2to1, resUnlocks[0])
	require.Equal(t, unlock2to2, resUnlocks[1])

	// delete all the records from creator 2
	keeper.RemoveLockup(ctx, unlock2to1)
	keeper.RemoveLockup(ctx, unlock2to2)
	_, found = keeper.GetLockup(ctx, addrCreators[1], denoms[0])
	require.False(t, found)
	_, found = keeper.GetLockup(ctx, addrCreators[1], denoms[1])
	require.False(t, found)
	resUnlocks = keeper.GetCreatorLockups(ctx, addrCreators[1], 5)
	require.Equal(t, 0, len(resUnlocks))
}

// func TestUnlockingRecord(t *testing.T) {
// 	keeper, ctx := keepertest.LiquidgovKeeper(t)

// 	addrCreators := createIncrementalAccounts(3)

// 	// construct the denoms
// 	var denoms [3]string
// 	denoms[0] = denom1
// 	denoms[1] = denom2
// 	denoms[2] = denom3

// 	ubd := types.NewUnlockingRecord(
// 		addrCreators[0],
// 		denoms[0],
// 		0,
// 		time.Unix(0, 0).UTC(),
// 		sdk.NewInt(5),
// 	)

// 	// set and retrieve a record
// 	keeper.SetUnlockingRecord(ctx, ubd)
// 	resUnbond, found := keeper.GetUnlockingRecord(ctx, addrCreators[0], denoms[0])
// 	require.True(t, found)
// 	require.Equal(t, ubd, resUnbond)

// 	// modify a records, save, and retrieve
// 	expUnbond := sdk.NewInt(21)
// 	ubd.Entries[0].Balance = expUnbond
// 	keeper.SetUnlockingRecord(ctx, ubd)

// 	resUnbonds := keeper.GetUnlockingRecords(ctx, addrCreators[0], 5)
// 	require.Equal(t, 1, len(resUnbonds))

// 	resUnbonds = keeper.GetAllUnlockingRecords(ctx, addrCreators[0])
// 	require.Equal(t, 1, len(resUnbonds))

// 	resUnbond, found = keeper.GetUnlockingRecord(ctx, addrCreators[0], denoms[0])
// 	require.True(t, found)
// 	require.Equal(t, ubd, resUnbond)

// 	resDelUnbond := keeper.GetDelegatorUnbonding(ctx, addrCreators[0])
// 	require.Equal(t, expUnbond, resDelUnbond)

// 	// delete a record
// 	keeper.RemoveUnlockingRecord(ctx, ubd)
// 	_, found = keeper.GetUnlockingRecord(ctx, addrCreators[0], denoms[0])
// 	require.False(t, found)

// 	resUnbonds = keeper.GetUnlockingRecords(ctx, addrCreators[0], 5)
// 	require.Equal(t, 0, len(resUnbonds))

// 	resUnbonds = keeper.GetAllUnlockingRecords(ctx, addrCreators[0])
// 	require.Equal(t, 0, len(resUnbonds))
// }
