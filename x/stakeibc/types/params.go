package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// Default init params
var (
	// these are default intervals _in epochs_ NOT in blocks
	DefaultDepositInterval        uint64 = 1
	DefaultDelegateInterval       uint64 = 1
	DefaultReinvestInterval       uint64 = 1
	DefaultRewardsInterval        uint64 = 1
	DefaultRedemptionRateInterval uint64 = 1
	// you apparently cannot safely encode floats, so we make commission / 100
	DefaultStrideCommission                 uint64 = 10
	DefaultValidatorRebalancingThreshold    uint64 = 100 // divide by 10,000, so 100 = 1%
	DefaultICATimeoutNanos                  uint64 = 600000000000
	DefaultBufferSize                       uint64 = 5             // 1/5=20% of the epoch
	DefaultIbcTimeoutBlocks                 uint64 = 300           // 300 blocks ~= 30 minutes
	DefaultFeeTransferTimeoutNanos          uint64 = 1800000000000 // 30 minutes
	DefaultSafetyMinRedemptionRateThreshold uint64 = 90            // divide by 100, so 90 = 0.9
	DefaultSafetyMaxRedemptionRateThreshold uint64 = 150           // divide by 100, so 150 = 1.5
	DefaultMaxStakeICACallsPerEpoch         uint64 = 100
	DefaultIBCTransferTimeoutNanos          uint64 = 1800000000000 // 30 minutes
	DefaultSafetyNumValidators              uint64 = 35
	DefaultSafetyMaxSlashPercent            uint64 = 10
	DefaultMaxRedemptionRates                      = make(map[string]string)
	DefaultMinRedemptionRates                      = make(map[string]string)

	// KeyDepositInterval is store's key for the DepositInterval option
	KeyDepositInterval                  = []byte("DepositInterval")
	KeyDelegateInterval                 = []byte("DelegateInterval")
	KeyReinvestInterval                 = []byte("ReinvestInterval")
	KeyRewardsInterval                  = []byte("RewardsInterval")
	KeyRedemptionRateInterval           = []byte("RedemptionRateInterval")
	KeyStrideCommission                 = []byte("StrideCommission")
	KeyValidatorRebalancingThreshold    = []byte("ValidatorRebalancingThreshold")
	KeyICATimeoutNanos                  = []byte("ICATimeoutNanos")
	KeyFeeTransferTimeoutNanos          = []byte("FeeTransferTimeoutNanos")
	KeyBufferSize                       = []byte("BufferSize")
	KeyIbcTimeoutBlocks                 = []byte("IBCTimeoutBlocks")
	KeySafetyMinRedemptionRateThreshold = []byte("SafetyMinRedemptionRateThreshold")
	KeySafetyMaxRedemptionRateThreshold = []byte("SafetyMaxRedemptionRateThreshold")
	KeyMaxStakeICACallsPerEpoch         = []byte("MaxStakeICACallsPerEpoch")
	KeyIBCTransferTimeoutNanos          = []byte("IBCTransferTimeoutNanos")
	KeySafetyNumValidators              = []byte("SafetyNumValidators")
	KeySafetyMaxSlashPercent            = []byte("SafetyMaxSlashPercent")
	KeyMaxRedemptionRates               = []byte("MaxRedemptionRates")
	KeyMinRedemptionRates               = []byte("MinRedemptionRates")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	depositInterval uint64,
	delegateInterval uint64,
	rewardsInterval uint64,
	redemptionRateInterval uint64,
	strideCommission uint64,
	reinvestInterval uint64,
	validatorRebalancingThreshold uint64,
	icaTimeoutNanos uint64,
	bufferSize uint64,
	ibcTimeoutBlocks uint64,
	feeTransferTimeoutNanos uint64,
	maxStakeIcaCallsPerEpoch uint64,
	safetyMinRedemptionRateThreshold uint64,
	safetyMaxRedemptionRateThreshold uint64,
	ibcTransferTimeoutNanos uint64,
	safetyNumValidators uint64,
	safetyMaxSlashPercent uint64,
	maxRedemptionRates map[string]string,
	minRedemptionRates map[string]string,
) Params {
	return Params{
		DepositInterval:                  depositInterval,
		DelegateInterval:                 delegateInterval,
		RewardsInterval:                  rewardsInterval,
		RedemptionRateInterval:           redemptionRateInterval,
		StrideCommission:                 strideCommission,
		ReinvestInterval:                 reinvestInterval,
		ValidatorRebalancingThreshold:    validatorRebalancingThreshold,
		IcaTimeoutNanos:                  icaTimeoutNanos,
		BufferSize:                       bufferSize,
		IbcTimeoutBlocks:                 ibcTimeoutBlocks,
		FeeTransferTimeoutNanos:          feeTransferTimeoutNanos,
		MaxStakeIcaCallsPerEpoch:         maxStakeIcaCallsPerEpoch,
		SafetyMinRedemptionRateThreshold: safetyMinRedemptionRateThreshold,
		SafetyMaxRedemptionRateThreshold: safetyMaxRedemptionRateThreshold,
		IbcTransferTimeoutNanos:          ibcTransferTimeoutNanos,
		SafetyNumValidators:              safetyNumValidators,
		SafetyMaxSlashPercent:            safetyMaxSlashPercent,
		MaxRedemptionRates:               maxRedemptionRates,
		MinRedemptionRates:               minRedemptionRates,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultDepositInterval,
		DefaultDelegateInterval,
		DefaultRewardsInterval,
		DefaultRedemptionRateInterval,
		DefaultStrideCommission,
		DefaultReinvestInterval,
		DefaultValidatorRebalancingThreshold,
		DefaultICATimeoutNanos,
		DefaultBufferSize,
		DefaultIbcTimeoutBlocks,
		DefaultFeeTransferTimeoutNanos,
		DefaultMaxStakeICACallsPerEpoch,
		DefaultSafetyMinRedemptionRateThreshold,
		DefaultSafetyMaxRedemptionRateThreshold,
		DefaultIBCTransferTimeoutNanos,
		DefaultSafetyNumValidators,
		DefaultSafetyMaxSlashPercent,
		DefaultMaxRedemptionRates,
		DefaultMinRedemptionRates,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDepositInterval, &p.DepositInterval, isPositive),
		paramtypes.NewParamSetPair(KeyDelegateInterval, &p.DelegateInterval, isPositive),
		paramtypes.NewParamSetPair(KeyRewardsInterval, &p.RewardsInterval, isPositive),
		paramtypes.NewParamSetPair(KeyRedemptionRateInterval, &p.RedemptionRateInterval, isPositive),
		paramtypes.NewParamSetPair(KeyStrideCommission, &p.StrideCommission, isCommission),
		paramtypes.NewParamSetPair(KeyReinvestInterval, &p.ReinvestInterval, isPositive),
		paramtypes.NewParamSetPair(KeyValidatorRebalancingThreshold, &p.ValidatorRebalancingThreshold, isThreshold),
		paramtypes.NewParamSetPair(KeyICATimeoutNanos, &p.IcaTimeoutNanos, isPositive),
		paramtypes.NewParamSetPair(KeyBufferSize, &p.BufferSize, isPositive),
		paramtypes.NewParamSetPair(KeyIbcTimeoutBlocks, &p.IbcTimeoutBlocks, isPositive),
		paramtypes.NewParamSetPair(KeyFeeTransferTimeoutNanos, &p.FeeTransferTimeoutNanos, validTimeoutNanos),
		paramtypes.NewParamSetPair(KeyMaxStakeICACallsPerEpoch, &p.MaxStakeIcaCallsPerEpoch, isPositive),
		paramtypes.NewParamSetPair(KeySafetyMinRedemptionRateThreshold, &p.SafetyMinRedemptionRateThreshold, validMinRedemptionRateThreshold),
		paramtypes.NewParamSetPair(KeySafetyMaxRedemptionRateThreshold, &p.SafetyMaxRedemptionRateThreshold, validMaxRedemptionRateThreshold),
		paramtypes.NewParamSetPair(KeyIBCTransferTimeoutNanos, &p.IbcTransferTimeoutNanos, validTimeoutNanos),
		paramtypes.NewParamSetPair(KeySafetyNumValidators, &p.SafetyNumValidators, isPositive),
		paramtypes.NewParamSetPair(KeySafetyMaxSlashPercent, &p.SafetyMaxSlashPercent, validSlashPercent),
		paramtypes.NewParamSetPair(KeyMaxRedemptionRates, &p.MaxRedemptionRates, validMaxRedemptionRates),
		paramtypes.NewParamSetPair(KeyMinRedemptionRates, &p.MinRedemptionRates, validMinRedemptionRates),
	}
}

func isThreshold(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("parameter not accepted: %T", i)
	}

	if ival <= 0 {
		return fmt.Errorf("parameter must be positive: %d", ival)
	}
	if ival > 10000 {
		return fmt.Errorf("parameter must be less than 10,000: %d", ival)
	}
	return nil
}

func validTimeoutNanos(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("parameter not accepted: %T", i)
	}

	tenMin := uint64(600000000000)
	oneHour := uint64(600000000000 * 6)

	if ival < tenMin {
		return fmt.Errorf("parameter must be g.t. 600000000000ns: %d", ival)
	}
	if ival > oneHour {
		return fmt.Errorf("parameter must be less than %dns: %d", oneHour, ival)
	}
	return nil
}

func validMaxRedemptionRateThreshold(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("parameter not accepted: %T", i)
	}

	maxVal := uint64(1000) // divide by 100, so 1000 => 10

	if ival > maxVal {
		return fmt.Errorf("parameter must be l.t. 1000: %d", ival)
	}

	return nil
}

func validMinRedemptionRateThreshold(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("parameter not accepted: %T", i)
	}

	minVal := uint64(75) // divide by 100, so 75 => 0.75

	if ival < minVal {
		return fmt.Errorf("parameter must be g.t. 75: %d", ival)
	}

	return nil
}

func validSlashPercent(i interface{}) error {
	ival, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("parameter not accepted: %T", i)
	}
	if ival > 100 {
		return fmt.Errorf("parameter must be between 0 and 100: %d", ival)
	}

	return nil
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

	if ival > 100 {
		return fmt.Errorf("commission must be less than 100: %d", ival)
	}
	return nil
}

func validMaxRedemptionRates(i interface{}) error {
	rrs, ok := i.(map[string]string)
	if !ok {
		return fmt.Errorf("max redemption rates not accepted: %T", i)
	}

	for _, rateStr := range rrs {
		rate, err := sdk.NewDecFromStr(rateStr)
		if err != nil {
			return err
		}
		if !rate.IsPositive() {
			return fmt.Errorf("redemption rate should be positive")
		}
	}

	return nil
}

func validMinRedemptionRates(i interface{}) error {
	rrs, ok := i.(map[string]string)
	if !ok {
		return fmt.Errorf("min redemption rates not accepted: %T", i)
	}

	for _, rateStr := range rrs {
		rate, err := sdk.NewDecFromStr(rateStr)
		if err != nil {
			return err
		}
		if !rate.IsPositive() {
			return fmt.Errorf("redemption rate should be positive")
		}
	}
	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := isPositive(p.DepositInterval); err != nil {
		return err
	}
	if err := isPositive(p.DelegateInterval); err != nil {
		return err
	}
	if err := isPositive(p.RewardsInterval); err != nil {
		return err
	}
	if err := isPositive(p.RedemptionRateInterval); err != nil {
		return err
	}
	if err := isCommission(p.StrideCommission); err != nil {
		return err
	}
	if err := isPositive(p.ReinvestInterval); err != nil {
		return err
	}
	if err := isThreshold(p.ValidatorRebalancingThreshold); err != nil {
		return err
	}
	if err := isPositive(p.IcaTimeoutNanos); err != nil {
		return err
	}
	if err := isPositive(p.BufferSize); err != nil {
		return err
	}
	if err := isPositive(p.IbcTimeoutBlocks); err != nil {
		return err
	}
	if err := validTimeoutNanos(p.FeeTransferTimeoutNanos); err != nil {
		return err
	}
	if err := isPositive(p.MaxStakeIcaCallsPerEpoch); err != nil {
		return err
	}
	if err := validMinRedemptionRateThreshold(p.SafetyMinRedemptionRateThreshold); err != nil {
		return err
	}
	if err := validMaxRedemptionRateThreshold(p.SafetyMaxRedemptionRateThreshold); err != nil {
		return err
	}
	if err := validTimeoutNanos(p.IbcTransferTimeoutNanos); err != nil {
		return err
	}
	if err := isPositive(p.SafetyNumValidators); err != nil {
		return err
	}
	if err := validSlashPercent(p.SafetyMaxSlashPercent); err != nil {
		return err
	}
	if err := validMaxRedemptionRates(p.MaxRedemptionRates); err != nil {
		return err
	}
	if err := validMinRedemptionRates(p.MinRedemptionRates); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
