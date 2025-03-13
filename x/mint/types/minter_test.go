package types

import (
	"math/rand"
	"testing"
)

// Benchmarking :)
// previously using math.Int operations:
// BenchmarkEpochProvision-4 5000000 220 ns/op
//
// using sdkmath.LegacyDec operations: (current implementation)
// BenchmarkEpochProvision-4 3000000 429 ns/op
func BenchmarkEpochProvision(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter()
	params := DefaultParams()

	s1 := rand.NewSource(100)
	r1 := rand.New(s1)
	minter.EpochProvisions = sdkmath.LegacyNewDec(r1.Int63n(1000000))

	// run the EpochProvision function b.N times
	for n := 0; n < b.N; n++ {
		minter.EpochProvision(params)
	}
}

// Next epoch provisions benchmarking
// BenchmarkNextEpochProvisions-4 5000000 251 ns/op
func BenchmarkNextEpochProvisions(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter()
	params := DefaultParams()

	// run the NextEpochProvisions function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextEpochProvisions(params)
	}
}
