package v2

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	oldclaimtypes "github.com/Stride-Labs/stride/v5/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v5/x/claim/types"
)

func TestConvertToNewAirdrop(t *testing.T) {
	id := "id1"
	startTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)
	duration := time.Duration(1)
	denom := "denom1"
	address := "address1"

	// Only the ClaimedSoFar field of the Airdrop should change
	oldAirdrop := oldclaimtypes.Airdrop{
		AirdropIdentifier:  id,
		AirdropStartTime:   startTime,
		AirdropDuration:    duration,
		ClaimDenom:         denom,
		DistributorAddress: address,
		ClaimedSoFar:       1,
	}
	expectedNewAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  id,
		AirdropStartTime:   startTime,
		AirdropDuration:    duration,
		ClaimDenom:         denom,
		DistributorAddress: address,
		ClaimedSoFar:       sdkmath.NewInt(1),
	}
	actualNewAirdrop := convertToNewAirdrop(oldAirdrop)
	require.Equal(t, expectedNewAirdrop, actualNewAirdrop)
}

func TestConvertToNewClaimParams(t *testing.T) {
	numAirdrops := 3

	// Build a list of old airdrops as well as the new expected type
	oldParams := oldclaimtypes.Params{}
	expectedNewParams := claimtypes.Params{}
	for i := 0; i <= numAirdrops-1; i++ {
		id := fmt.Sprintf("id-%d", i)

		oldParams.Airdrops = append(oldParams.Airdrops, &oldclaimtypes.Airdrop{
			AirdropIdentifier: id,
			ClaimedSoFar:      int64(i),
		})

		expectedNewParams.Airdrops = append(expectedNewParams.Airdrops, &claimtypes.Airdrop{
			AirdropIdentifier: id,
			ClaimedSoFar:      sdk.NewInt(int64(i)),
		})
	}

	// Convert airdrop params
	actualNewParams := convertToNewClaimParams(oldParams)

	// Confirm new params align with expectations
	require.Equal(t, len(expectedNewParams.Airdrops), len(actualNewParams.Airdrops))
	for i := 0; i <= numAirdrops-1; i++ {
		require.Equal(t, expectedNewParams.Airdrops[i], actualNewParams.Airdrops[i], "index: %d", i)
	}
}
