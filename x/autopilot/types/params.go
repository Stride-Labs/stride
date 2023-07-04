package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	// Default active value for each autopilot supported module
	DefaultStakeibcActive = true
	DefaultClaimActive    = true
)

// KeyActive is the store key for Params
var KeyStakeibcActive = []byte("StakeibcActive")
var KeyClaimActive = []byte("ClaimActive")

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(stakeibcActive, claimActive bool) Params {
	return Params{
		StakeibcActive: stakeibcActive,
		ClaimActive:    claimActive,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultStakeibcActive, DefaultClaimActive)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyStakeibcActive, &p.StakeibcActive, validateBool),
		paramtypes.NewParamSetPair(KeyClaimActive, &p.ClaimActive, validateBool),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateBool(p.StakeibcActive); err != nil {
		return err
	}
	if err := validateBool(p.ClaimActive); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
