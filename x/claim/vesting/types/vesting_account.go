package types

import (
	"errors"
	"time"

	yaml "gopkg.in/yaml.v2"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/v4/utils"
	vestexported "github.com/Stride-Labs/stride/v4/x/claim/vesting/exported"
)

// Compile-time type assertions
var (
	_ authtypes.AccountI          = (*BaseVestingAccount)(nil)
	_ vestexported.VestingAccount = (*StridePeriodicVestingAccount)(nil)
)

// Base Vesting Account

// NewBaseVestingAccount creates a new BaseVestingAccount object. It is the
// callers responsibility to ensure the base account has sufficient funds with
// regards to the original vesting amount.
func NewBaseVestingAccount(baseAccount *authtypes.BaseAccount, originalVesting sdk.Coins, endTime int64) *BaseVestingAccount {
	return &BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  originalVesting,
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		EndTime:          endTime,
	}
}

// LockedCoinsFromVesting returns all the coins that are not spendable (i.e. locked)
// for a vesting account given the current vesting coins. If no coins are locked,
// an empty slice of Coins is returned.
//
// CONTRACT: Delegated vesting coins and vestingCoins must be sorted.
func (bva BaseVestingAccount) LockedCoinsFromVesting(vestingCoins sdk.Coins) sdk.Coins {
	lockedCoins := vestingCoins.Sub(vestingCoins.Min(bva.DelegatedVesting))
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVestingAccount) TrackDelegation(balance, vestingCoins, amount sdk.Coins) {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt := bva.DelegatedVesting.AmountOf(coin.Denom)

		// Panic if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			panic("delegation attempt with zero coins or insufficient funds")
		}

		// compute x and y per the specification, where:
		// X := min(max(V - DV, 0), D)
		// Y := D - X
		x := sdk.MinInt(sdk.MaxInt(vestingAmt.Sub(delVestingAmt), sdk.ZeroInt()), coin.Amount)
		y := coin.Amount.Sub(x)

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedVesting = bva.DelegatedVesting.Add(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedFree = bva.DelegatedFree.Add(yCoin)
		}
	}
}

// TrackUndelegation tracks an undelegation amount by setting the necessary
// values by which delegated vesting and delegated vesting need to decrease and
// by which amount the base coins need to increase.
//
// NOTE: The undelegation (bond refund) amount may exceed the delegated
// vesting (bond) amount due to the way undelegation truncates the bond refund,
// which can increase the validator's exchange rate (tokens/shares) slightly if
// the undelegated tokens are non-integral.
//
// CONTRACT: The account's coins and undelegation coins must be sorted.
func (bva *BaseVestingAccount) TrackUndelegation(amount sdk.Coins) {
	for _, coin := range amount {
		// panic if the undelegation amount is zero
		if coin.Amount.IsZero() {
			panic("undelegation attempt with zero coins")
		}
		delegatedFree := bva.DelegatedFree.AmountOf(coin.Denom)
		delegatedVesting := bva.DelegatedVesting.AmountOf(coin.Denom)

		// compute x and y per the specification, where:
		// X := min(DF, D)
		// Y := min(DV, D - X)
		x := sdk.MinInt(delegatedFree, coin.Amount)
		y := sdk.MinInt(delegatedVesting, coin.Amount.Sub(x))

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedFree = bva.DelegatedFree.Sub(sdk.Coins{xCoin})
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedVesting = bva.DelegatedVesting.Sub(sdk.Coins{yCoin})
		}
	}
}

// GetOriginalVesting returns a vesting account's original vesting amount
func (bva BaseVestingAccount) GetOriginalVesting() sdk.Coins {
	return bva.OriginalVesting
}

// GetDelegatedFree returns a vesting account's delegation amount that is not
// vesting.
func (bva BaseVestingAccount) GetDelegatedFree() sdk.Coins {
	return bva.DelegatedFree
}

// GetDelegatedVesting returns a vesting account's delegation amount that is
// still vesting.
func (bva BaseVestingAccount) GetDelegatedVesting() sdk.Coins {
	return bva.DelegatedVesting
}

// GetEndTime returns a vesting account's end time
func (bva BaseVestingAccount) GetEndTime() int64 {
	return bva.EndTime
}

// Validate checks for errors on the account fields
func (bva BaseVestingAccount) Validate() error {
	if !(bva.DelegatedVesting.IsAllLTE(bva.OriginalVesting)) {
		return errors.New("delegated vesting amount cannot be greater than original vesting amount")
	}
	return bva.BaseAccount.Validate()
}

type vestingAccountYAML struct {
	Address          sdk.AccAddress `json:"address" yaml:"address"`
	PubKey           string         `json:"public_key" yaml:"public_key"`
	AccountNumber    uint64         `json:"account_number" yaml:"account_number"`
	Sequence         uint64         `json:"sequence" yaml:"sequence"`
	OriginalVesting  sdk.Coins      `json:"original_vesting" yaml:"original_vesting"`
	DelegatedFree    sdk.Coins      `json:"delegated_free" yaml:"delegated_free"`
	DelegatedVesting sdk.Coins      `json:"delegated_vesting" yaml:"delegated_vesting"`
	EndTime          int64          `json:"end_time" yaml:"end_time"`

	// custom fields based on concrete vesting type which can be omitted
	StartTime      int64   `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	VestingPeriods Periods `json:"vesting_periods,omitempty" yaml:"vesting_periods,omitempty"`
}

func (bva BaseVestingAccount) String() string {
	out, _ := bva.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a BaseVestingAccount.
func (bva BaseVestingAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(bva.Address)
	if err != nil {
		return nil, err
	}

	out := vestingAccountYAML{
		Address:          accAddr,
		AccountNumber:    bva.AccountNumber,
		PubKey:           getPKString(bva),
		Sequence:         bva.Sequence,
		OriginalVesting:  bva.OriginalVesting,
		DelegatedFree:    bva.DelegatedFree,
		DelegatedVesting: bva.DelegatedVesting,
		EndTime:          bva.EndTime,
	}
	return marshalYaml(out)
}

// Periodic Vesting Account (only for stride)
// This vesting account works differently from the core periodic vesting account.
var _ vestexported.VestingAccount = (*StridePeriodicVestingAccount)(nil)
var _ authtypes.GenesisAccount = (*StridePeriodicVestingAccount)(nil)

// NewStridePeriodicVestingAccountRaw creates a new StridePeriodicVestingAccount object from BaseVestingAccount
func NewStridePeriodicVestingAccountRaw(bva *BaseVestingAccount, startTime int64, periods Periods) *StridePeriodicVestingAccount {
	return &StridePeriodicVestingAccount{
		BaseVestingAccount: bva,
		VestingPeriods:     periods,
	}
}

// NewStridePeriodicVestingAccount returns a new StridePeriodicVestingAccount
func NewStridePeriodicVestingAccount(baseAcc *authtypes.BaseAccount, originalVesting sdk.Coins, periods Periods) *StridePeriodicVestingAccount {
	if len(periods) == 0 {
		return &StridePeriodicVestingAccount{}
	}

	endTime := int64(0)
	for _, p := range periods {
		endTime = utils.Max64(endTime, p.StartTime+p.Length)
	}

	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	return &StridePeriodicVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		VestingPeriods:     periods,
	}
}

// AddNewGrant adds a new grant
func (pva *StridePeriodicVestingAccount) AddNewGrant(grantedPeriod Period) {
	// Starting time for new period must be greater than original starting time
	pva.VestingPeriods = append(pva.VestingPeriods, grantedPeriod)
	pva.EndTime = utils.Max64(pva.EndTime, grantedPeriod.Length+grantedPeriod.StartTime)
	pva.OriginalVesting = pva.OriginalVesting.Add(grantedPeriod.Amount...)
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva StridePeriodicVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if len(pva.VestingPeriods) == 0 {
		return vestedCoins
	} else if blockTime.Unix() <= pva.VestingPeriods[0].StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= pva.EndTime {
		return pva.OriginalVesting
	}

	for _, period := range pva.VestingPeriods {
		vestedCoins = vestedCoins.Add(utils.GetVestedCoinsAt(blockTime.Unix(), period.StartTime, period.Length, period.Amount)...)
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva StridePeriodicVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return pva.OriginalVesting.Sub(pva.GetVestedCoins(blockTime))
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (pva StridePeriodicVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return pva.BaseVestingAccount.LockedCoinsFromVesting(pva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (pva *StridePeriodicVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	pva.BaseVestingAccount.TrackDelegation(balance, pva.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns the time when vesting starts for a periodic vesting
// account.
func (pva StridePeriodicVestingAccount) GetStartTime() int64 {
	return pva.VestingPeriods[0].StartTime
}

// GetVestingPeriods returns vesting periods associated with periodic vesting account.
func (pva StridePeriodicVestingAccount) GetVestingPeriods() Periods {
	return pva.VestingPeriods
}

// Validate checks for errors on the account fields
func (pva StridePeriodicVestingAccount) Validate() error {
	if pva.GetStartTime() >= pva.GetEndTime() {
		return errors.New("vesting start-time cannot be before end-time")
	}
	endTime := pva.VestingPeriods[0].StartTime
	originalVesting := sdk.NewCoins()
	for _, p := range pva.VestingPeriods {
		endTime += p.Length
		originalVesting = originalVesting.Add(p.Amount...)
	}
	if endTime != pva.EndTime {
		return errors.New("vesting end time does not match length of all vesting periods")
	}
	if !originalVesting.IsEqual(pva.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return pva.BaseVestingAccount.Validate()
}

func (pva StridePeriodicVestingAccount) String() string {
	out, _ := pva.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a StridePeriodicVestingAccount.
func (pva StridePeriodicVestingAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(pva.Address)
	if err != nil {
		return nil, err
	}

	out := vestingAccountYAML{
		Address:          accAddr,
		AccountNumber:    pva.AccountNumber,
		PubKey:           getPKString(pva),
		Sequence:         pva.Sequence,
		OriginalVesting:  pva.OriginalVesting,
		DelegatedFree:    pva.DelegatedFree,
		DelegatedVesting: pva.DelegatedVesting,
		EndTime:          pva.EndTime,
		StartTime:        pva.VestingPeriods[0].StartTime,
		VestingPeriods:   pva.VestingPeriods,
	}
	return marshalYaml(out)
}

type getPK interface {
	GetPubKey() cryptotypes.PubKey
}

func getPKString(g getPK) string {
	if pk := g.GetPubKey(); pk != nil {
		return pk.String()
	}
	return ""
}

func marshalYaml(i interface{}) (interface{}, error) {
	bz, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}
	return string(bz), nil
}
