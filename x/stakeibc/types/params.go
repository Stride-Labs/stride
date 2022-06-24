package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// Default init params
var (
	// these are default intervals _in epochs_ NOT in blocks
	DefaultDepositInterval       uint64 = 3
	DefaultDelegateInterval      uint64 = 3
	DefaultReinvestInterval      uint64 = 3
	DefaultRewardsInterval       uint64 = 3
	DefaultExchangeRateInterval  uint64 = 3
	DefaultKeyWithdrawalInterval uint64 = 3
	// you apparantly cannot safely encode floats, so we make commission / 100
	DefaultStrideCommission uint64 = 10

	// KeyDepositInterval is store's key for the DepositInterval option
	KeyDepositInterval      = []byte("DepositInterval")
	KeyDelegateInterval     = []byte("DelegateInterval")
	KeyReinvestInterval     = []byte("ReinvestInterval")
	KeyRewardsInterval      = []byte("RewardsInterval")
	KeyExchangeRateInterval = []byte("ExchangeRateInterval")
	KeyStrideCommission     = []byte("StrideCommission")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	deposit_interval uint64,
	delegate_interval uint64,
	rewards_interval uint64,
	exchange_rate_interval uint64,
	stride_commission uint64,
	reinvest_interval uint64,
) Params {
	return Params{
		DepositInterval:      deposit_interval,
		DelegateInterval:     delegate_interval,
		RewardsInterval:      rewards_interval,
		ExchangeRateInterval: exchange_rate_interval,
		StrideCommission:     stride_commission,
		ReinvestInterval:     reinvest_interval,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultDepositInterval,
		DefaultDelegateInterval,
		DefaultRewardsInterval,
		DefaultExchangeRateInterval,
		DefaultStrideCommission,
		DefaultReinvestInterval,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDepositInterval, &p.DepositInterval, isPositive),
		paramtypes.NewParamSetPair(KeyDelegateInterval, &p.DelegateInterval, isPositive),
		paramtypes.NewParamSetPair(KeyRewardsInterval, &p.RewardsInterval, isPositive),
		paramtypes.NewParamSetPair(KeyExchangeRateInterval, &p.ExchangeRateInterval, isPositive),
		paramtypes.NewParamSetPair(KeyStrideCommission, &p.StrideCommission, isCommission),
		paramtypes.NewParamSetPair(KeyReinvestInterval, &p.ReinvestInterval, isPositive),
	}
}

func isPositive(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("parameter not accepted: %T", i)
	}

	if ival <= 0 {
		return fmt.Errorf("parameter must be positive: %d", ival)
	}
	return nil
}

func isCommission(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("commission not accepted: %T", i)
	}

	if ival < 0 {
		return fmt.Errorf("commission must be non-negative: %d", ival)
	} else if ival > 100 {
		return fmt.Errorf("commission must be less than 100: %d", ival)
	}
	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
