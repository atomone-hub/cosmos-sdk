package keeper

import (
	"sort"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AdjustEta is called to adjust η dynamically for each block.
func (k Keeper) AdjustEta(ctx sdk.Context) error {
	if ctx.BlockHeight()%types.NakamotoBonusUpdateInterval != 0 {
		return nil
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	if !params.NakamotoBonusEnabled {
		// Always set eta to zero and skip dynamic update
		if params.NakamotoBonusCoefficient.IsZero() {
			// Already zero, nothing to do
			return nil
		}
		params.NakamotoBonusCoefficient = math.LegacyZeroDec()
		return k.Params.Set(ctx, params)
	}

	validators, err := k.stakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return err
	}
	n := len(validators)
	if n < 3 {
		return nil // Not enough validators to split into three groups
	}
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].GetBondedTokens().GT(validators[j].GetBondedTokens())
	})

	// Dynamically divide into three groups (high, medium, low) as evenly as possible
	// high: first groupSize, medium: next groupSize, low: rest
	groupSize := n / 3

	highEnd := groupSize
	mediumEnd := groupSize * 2

	// If there is a remainder, distribute it to the last group ("low")
	// So low group will have groupSize + remainder
	lowStart := mediumEnd
	lowEnd := n

	var high, low []stakingtypes.Validator
	high = validators[:highEnd]
	low = validators[lowStart:lowEnd]

	sum := func(vals []stakingtypes.Validator) math.Int {
		total := math.ZeroInt()
		for _, v := range vals {
			total = total.Add(v.GetBondedTokens())
		}
		return total
	}
	avg := func(vals []stakingtypes.Validator) math.LegacyDec {
		if len(vals) == 0 {
			return math.LegacyZeroDec()
		}
		return math.LegacyNewDecFromInt(sum(vals)).QuoInt64(int64(len(vals)))
	}
	highAvg := avg(high)
	lowAvg := avg(low)
	coefficient := params.NakamotoBonusCoefficient

	// Adjust coefficient: if avgHigh >= 3x avgLow, increase eta, else decrease
	// NakamotoBonusStep should be a decimal value, e.g. 0.03 for 3%
	if lowAvg.IsZero() || highAvg.Quo(lowAvg).GTE(math.LegacyNewDec(types.NakamotoBonusStep)) {
		coefficient = coefficient.Add(math.LegacyNewDecWithPrec(types.NakamotoBonusStep, 2))
	} else {
		coefficient = coefficient.Sub(math.LegacyNewDecWithPrec(types.NakamotoBonusStep, 2))
	}
	if coefficient.LT(math.LegacyZeroDec()) {
		coefficient = math.LegacyZeroDec()
	}
	if coefficient.GT(math.LegacyOneDec()) {
		coefficient = math.LegacyOneDec()
	}

	if !coefficient.Equal(params.NakamotoBonusCoefficient) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeNakamotoCoefficient,
				sdk.NewAttribute(types.AttributeNakamotoCoefficient, coefficient.String()),
			),
		)
	}

	params.NakamotoBonusCoefficient = coefficient
	return k.Params.Set(ctx, params)
}
