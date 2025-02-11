package gamm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/deps/osmomath"
)

var (
	_ PoolI     = &OsmosisGammPool{}
	_ CFMMPoolI = &OsmosisGammPool{}
)

// PoolI defines an interface for pools that hold tokens.
type PoolI interface {
	proto.Message

	GetAddress() sdk.AccAddress
	String() string
	GetId() uint64
	// GetSpreadFactor returns the pool's spread factor, based on the current state.
	// Pools may choose to make their spread factors dependent upon state
	// (prior TWAPs, network downtime, other pool states, etc.)
	// hence Context is provided as an argument.
	GetSpreadFactor(ctx sdk.Context) osmomath.Dec
	// Returns whether the pool has swaps enabled at the moment
	IsActive(ctx sdk.Context) bool

	// GetPoolDenoms returns the pool's denoms.
	GetPoolDenoms(sdk.Context) []string

	// Returns the spot price of the 'base asset' in terms of the 'quote asset' in the pool,
	// errors if either baseAssetDenom, or quoteAssetDenom does not exist.
	// For example, if this was a UniV2 50-50 pool, with 2 ETH, and 8000 UST
	// pool.SpotPrice(ctx, "eth", "ust") = 4000.00
	SpotPrice(ctx sdk.Context, quoteAssetDenom string, baseAssetDenom string) (osmomath.BigDec, error)
	// GetType returns the type of the pool (Balancer, Stableswap, Concentrated, etc.)
	GetType() PoolType
	// AsSerializablePool returns the pool in a serializable form (useful when a model wraps the proto)
	AsSerializablePool() PoolI
}

// PoolType is an enumeration of all supported pool types.
type PoolType int32

const (
	// Balancer is the standard xy=k curve. Its pool model is defined in x/gamm.
	Balancer PoolType = 0
	// Stableswap is the Solidly cfmm stable swap curve. Its pool model is defined
	// in x/gamm.
	Stableswap PoolType = 1
	// Concentrated is the pool model specific to concentrated liquidity. It is
	// defined in x/concentrated-liquidity.
	Concentrated PoolType = 2
	// CosmWasm is the pool model specific to CosmWasm. It is defined in
	// x/cosmwasmpool.
	CosmWasm PoolType = 3
)

func (p *OsmosisGammPool) AsSerializablePool() PoolI {
	panic("AsSerializablePool should not be called")
}

// GetAddress returns the address of a pool.
// If the pool address is not bech32 valid, it returns an empty address.
func (p OsmosisGammPool) GetAddress() sdk.AccAddress {
	panic("GetAddress should not be called")
}

func (p OsmosisGammPool) GetId() uint64 {
	panic("GetId should not be called")
}

func (p OsmosisGammPool) GetSpreadFactor(_ sdk.Context) osmomath.Dec {
	panic("GetSpreadFactor should not be called")
}

// GetPoolDenoms implements types.CFMMPoolI.
func (p *OsmosisGammPool) GetPoolDenoms(ctx sdk.Context) []string {
	panic("GetPoolDenoms should not be called")
}

func (p OsmosisGammPool) GetType() PoolType {
	panic("GetType should not be called")
}

func (p OsmosisGammPool) IsActive(ctx sdk.Context) bool {
	panic("IsActive should not be called")
}

func (p OsmosisGammPool) SpotPrice(ctx sdk.Context, quoteAssetDenom string, baseAssetDenom string) (osmomath.BigDec, error) {
	panic("SpotPrice should not be called")
}

// CFMMPoolI defines an interface for pools representing constant function
// AMM.
type CFMMPoolI interface {
	PoolI

	// JoinPool joins the pool using all of the tokensIn provided.
	// The AMM swaps to the correct internal ratio should be and returns the number of shares created.
	// This function is mutative and updates the pool's internal state if there is no error.
	// It is up to pool implementation if they support LP'ing at arbitrary ratios, or a subset of ratios.
	// Pools are expected to guarantee LP'ing at the exact ratio, and single sided LP'ing.
	JoinPool(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, err error)
	// JoinPoolNoSwap joins the pool with an all-asset join using the maximum amount possible given the tokensIn provided.
	// This function is mutative and updates the pool's internal state if there is no error.
	// Pools are expected to guarantee LP'ing at the exact ratio.
	JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, err error)

	// ExitPool exits #numShares LP shares from the pool, decreases its internal liquidity & LP share totals,
	// and returns the number of coins that are being returned.
	// This mutates the pool and state.
	ExitPool(ctx sdk.Context, numShares osmomath.Int, exitFee osmomath.Dec) (exitedCoins sdk.Coins, err error)
	// CalcJoinPoolNoSwapShares returns how many LP shares JoinPoolNoSwap would return on these arguments.
	// This does not mutate the pool, or state.
	CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, newLiquidity sdk.Coins, err error)
	// CalcExitPoolCoinsFromShares returns how many coins ExitPool would return on these arguments.
	// This does not mutate the pool, or state.
	CalcExitPoolCoinsFromShares(ctx sdk.Context, numShares osmomath.Int, exitFee osmomath.Dec) (exitedCoins sdk.Coins, err error)
	// CalcJoinPoolShares returns how many LP shares JoinPool would return on these arguments.
	// This does not mutate the pool, or state.
	CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, newLiquidity sdk.Coins, err error)
	// SwapOutAmtGivenIn swaps 'tokenIn' against the pool, for tokenOutDenom, with the provided spreadFactor charged.
	// Balance transfers are done in the keeper, but this method updates the internal pool state.
	SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, spreadFactor osmomath.Dec) (tokenOut sdk.Coin, err error)
	// CalcOutAmtGivenIn returns how many coins SwapOutAmtGivenIn would return on these arguments.
	// This does not mutate the pool, or state.
	CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, spreadFactor osmomath.Dec) (tokenOut sdk.Coin, err error)

	// SwapInAmtGivenOut swaps exactly enough tokensIn against the pool, to get the provided tokenOut amount out of the pool.
	// Balance transfers are done in the keeper, but this method updates the internal pool state.
	SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, spreadFactor osmomath.Dec) (tokenIn sdk.Coin, err error)
	// CalcInAmtGivenOut returns how many coins SwapInAmtGivenOut would return on these arguments.
	// This does not mutate the pool, or state.
	CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, spreadFactor osmomath.Dec) (tokenIn sdk.Coin, err error)
	// GetTotalShares returns the total number of LP shares in the pool
	GetTotalShares() osmomath.Int
	// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
	GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins
	// GetExitFee returns the pool's exit fee, based on the current state.
	// Pools may choose to make their exit fees dependent upon state.
	GetExitFee(ctx sdk.Context) osmomath.Dec
}

func (p OsmosisGammPool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, err error) {
	panic("JoinPool should not be called")
}

func (p OsmosisGammPool) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, err error) {
	panic("JoinPoolNoSwap should not be called")
}

func (p OsmosisGammPool) ExitPool(ctx sdk.Context, numShares osmomath.Int, exitFee osmomath.Dec) (exitedCoins sdk.Coins, err error) {
	panic("ExitPool should not be called")
}

func (p OsmosisGammPool) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, newLiquidity sdk.Coins, err error) {
	panic("CalcJoinPoolNoSwapShares should not be called")
}

func (p OsmosisGammPool) CalcExitPoolCoinsFromShares(ctx sdk.Context, numShares osmomath.Int, exitFee osmomath.Dec) (exitedCoins sdk.Coins, err error) {
	panic("CalcExitPoolCoinsFromShares should not be called")
}

func (p OsmosisGammPool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor osmomath.Dec) (numShares osmomath.Int, newLiquidity sdk.Coins, err error) {
	panic("CalcJoinPoolShares should not be called")
}

func (p OsmosisGammPool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, spreadFactor osmomath.Dec) (tokenOut sdk.Coin, err error) {
	panic("SwapOutAmtGivenIn should not be called")
}

func (p OsmosisGammPool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, spreadFactor osmomath.Dec) (tokenOut sdk.Coin, err error) {
	panic("CalcOutAmtGivenIn should not be called")
}

func (p OsmosisGammPool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, spreadFactor osmomath.Dec) (tokenIn sdk.Coin, err error) {
	panic("SwapInAmtGivenOut should not be called")
}

func (p OsmosisGammPool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, spreadFactor osmomath.Dec) (tokenIn sdk.Coin, err error) {
	panic("CalcInAmtGivenOut should not be called")
}

func (p OsmosisGammPool) GetTotalShares() osmomath.Int {
	panic("GetTotalShares should not be called")
}

func (p OsmosisGammPool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	panic("GetTotalPoolLiquidity should not be called")
}

func (p OsmosisGammPool) GetExitFee(ctx sdk.Context) osmomath.Dec {
	panic("GetExitFee should not be called")
}
