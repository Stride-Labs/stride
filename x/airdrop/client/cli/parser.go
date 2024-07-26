package cli

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v23/x/airdrop/types"
)

// Parses an allocations CSV file consisting of allocations for various addresses
//
// Example Schema:
//
//	strideXXX,10,10,20
//	strideYYY,0,10,0
func ParseUserAllocations(fileName string) ([]types.RawAllocation, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	allAllocations := []types.RawAllocation{}
	for _, row := range rows {
		if len(row) < 2 {
			return nil, errors.New("invalid csv row")
		}

		userAddress := row[0]

		allocations := []sdkmath.Int{}
		for _, allocationString := range row[1:] {
			allocation, ok := sdkmath.NewIntFromString(allocationString)
			if !ok {
				return nil, fmt.Errorf("unable to parse allocation %s into sdk.Int", allocationString)
			}
			allocations = append(allocations, allocation)
		}

		allocation := types.RawAllocation{
			UserAddress: userAddress,
			Allocations: allocations,
		}
		allAllocations = append(allAllocations, allocation)
	}

	return allAllocations, nil
}
