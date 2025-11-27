package types

import (
	"fmt"

	"cosmossdk.io/math"
)

const (
	// DefaultNakamotoBonusPeriod represents default nakamoto bonus period (in blocks)
	DefaultNakamotoBonusPeriod = 120_000
	// defaultNakamotoBonusStep represents the default step to increase or decrease η
	defaultNakamotoBonusStep = 3
)

var DefaultNakamotoBonusStep = math.LegacyNewDecWithPrec(defaultNakamotoBonusStep, 2)

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		BaseProposerReward:  math.LegacyZeroDec(),
		BonusProposerReward: math.LegacyZeroDec(),
		CommunityTax:        math.LegacyNewDecWithPrec(2, 2), // 2%
		WithdrawAddrEnabled: true,
		NakamotoBonus: NakamotoBonus{
			Enabled: true,
			Step:    DefaultNakamotoBonusStep,
			Period:  DefaultNakamotoBonusPeriod,
		},
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	if err := validateCommunityTax(p.CommunityTax); err != nil {
		return err
	}
	return validateNakamotoBonus(p.NakamotoBonus)
}

func validateCommunityTax(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	switch {
	case v.IsNil():
		return fmt.Errorf("community tax must be not nil")
	case v.IsNegative():
		return fmt.Errorf("community tax must be positive: %s", v)
	case v.GT(math.LegacyOneDec()):
		return fmt.Errorf("community tax too large: %s", v)
	}
	return nil
}

func validateWithdrawAddrEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateNakamotoBonus(i interface{}) error {
	v, ok := i.(NakamotoBonus)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Period == 0 {
		return fmt.Errorf("nakamoto bonus period must be greater than zero: %d", v.Period)
	}

	switch {
	case v.Step.IsNil():
		return fmt.Errorf("nakamoto bonus step must be not nil")
	case v.Step.IsNegative():
		return fmt.Errorf("nakamoto bonus step must be positive: %v", v.Step)
	case v.Step.GT(math.LegacyOneDec()):
		return fmt.Errorf("nakamoto bonus step too large: %v", v.Step)
	}
	return nil
}
