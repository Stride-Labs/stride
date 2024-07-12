package cli

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

// Parses an allocations CSV file consisting of allocations for various addresses
//
// Example Schema:
//
//	strideXXX,10,10,20
//	strideYYY,0,10,0
func ParseMultipleUserAllocations(fileName string) ([]types.RawAllocation, error) {
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

// Parses a single user's allocations from a single line file with comma separate reward amounts
// Ex: 10,10,20
func ParseSingleUserAllocations(fileName string) (allocations []sdkmath.Int, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var content string

	if scanner.Scan() {
		content = scanner.Text()
	}

	if scanner.Scan() {
		return nil, fmt.Errorf("file %s has more than one line", fileName)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	allocationsSplit := strings.Split(content, ",")
	for _, allocationString := range allocationsSplit {
		allocation, ok := sdkmath.NewIntFromString(allocationString)
		if !ok {
			return nil, errors.New("unable to parse reward")
		}
		allocations = append(allocations, allocation)
	}

	return allocations, nil
}
