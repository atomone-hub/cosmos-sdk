package types

import (
	"fmt"

	"cosmossdk.io/math"
)

const (
	// NakamotoBonusUpdateInterval represents 120k blocks
	NakamotoBonusUpdateInterval = 120_000
	// NakamotoBonusStep represents the step to increase or decrease η
	NakamotoBonusStep = 3
)

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		BaseProposerReward:       math.LegacyZeroDec(),
		BonusProposerReward:      math.LegacyZeroDec(),
		CommunityTax:             math.LegacyNewDecWithPrec(2, 2), // 2%
		WithdrawAddrEnabled:      true,
		NakamotoBonusCoefficient: math.LegacyNewDecWithPrec(NakamotoBonusStep, 2), // 3%
		NakamotoBonusEnabled:     true,
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	if err := validateCommunityTax(p.CommunityTax); err != nil {
		return err
	}
	return validateNakamotoBonusCoefficient(p.NakamotoBonusCoefficient)
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

func validateNakamotoBonusCoefficient(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	switch {
	case v.IsNil():
		return fmt.Errorf("nakamoto bonus coefficient must be not nil")
	case v.IsNegative():
		return fmt.Errorf("nakamoto bonus coefficient must be positive: %s", v)
	case v.GT(math.LegacyOneDec()):
		return fmt.Errorf("nakamoto bonus coefficient too large: %s", v)
	}
	return nil
}

func validateNakamotoBonusEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
