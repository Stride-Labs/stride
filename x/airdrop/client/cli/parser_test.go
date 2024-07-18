package cli_test

import (
	"os"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/x/airdrop/client/cli"
	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func ParseUserAllocations(t *testing.T) {
	inputCSVContents := `strideXXX,10,10,20
strideYYY,0,10,0
strideZZZ,5,100,6`

	expectedAllocations := []types.RawAllocation{
		{
			UserAddress: "strideXXX",
			Allocations: []sdkmath.Int{sdkmath.NewInt(10), sdkmath.NewInt(10), sdkmath.NewInt(20)},
		},
		{
			UserAddress: "strideYYY",
			Allocations: []sdkmath.Int{sdkmath.NewInt(0), sdkmath.NewInt(10), sdkmath.NewInt(0)},
		},
		{
			UserAddress: "strideZZZ",
			Allocations: []sdkmath.Int{sdkmath.NewInt(5), sdkmath.NewInt(100), sdkmath.NewInt(6)},
		},
	}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "allocations*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write the CSV string to the temp file
	_, err = tmpfile.WriteString(inputCSVContents)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	// Call the function with the temporary file name
	actualAllocations, err := cli.ParseUserAllocations(tmpfile.Name())
	require.NoError(t, err)

	// Validate the allocations match expectations
	require.Equal(t, expectedAllocations, actualAllocations)
}
