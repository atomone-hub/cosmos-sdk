package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func createValidators(powers ...int64) ([]stakingtypes.Validator, error) {
	vals := make([]stakingtypes.Validator, len(powers))
	for i, p := range powers {
		vals[i] = stakingtypes.Validator{
			OperatorAddress: sdk.ValAddress([]byte{byte(i)}).String(),
			Tokens:          math.NewInt(p),
			Status:          stakingtypes.Bonded,
			Commission:      stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
		}
	}
	return vals, nil
}

func TestAdjustEta_NakamotoDisabled(t *testing.T) {
	s := setupTestKeeper(t, types.DefaultNakamotoBonusStep, types.DefaultNakamotoBonusPeriod)

	p, err := s.distrKeeper.Params.Get(s.ctx)
	require.NoError(t, err)
	p.NakamotoBonus.Enabled = false
	require.NoError(t, s.distrKeeper.Params.Set(s.ctx, p))

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, math.LegacyZeroDec(), nakamotoBonusCoefficient)
}

func TestAdjustEta_NoInterval(t *testing.T) {
	s := setupTestKeeper(t, types.DefaultNakamotoBonusStep, 119_999)

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultNakamotoBonusStep, nakamotoBonusCoefficient)
}

func TestAdjustEta_NotEnoughValidators(t *testing.T) {
	s := setupTestKeeper(t, types.DefaultNakamotoBonusStep, types.DefaultNakamotoBonusPeriod)

	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(createValidators(10, 10)).AnyTimes()

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultNakamotoBonusStep, nakamotoBonusCoefficient)
}

func TestAdjustEta_Increase(t *testing.T) {
	s := setupTestKeeper(t, types.DefaultNakamotoBonusStep, types.DefaultNakamotoBonusPeriod)

	// highAvg = 100, lowAvg = 10, ratio = 10 >= 3, should increase
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(createValidators(100, 100, 10)).AnyTimes()

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultNakamotoBonusStep.Add(types.DefaultNakamotoBonusStep), nakamotoBonusCoefficient)
}

func TestAdjustEta_Decrease(t *testing.T) {
	s := setupTestKeeper(t, types.DefaultNakamotoBonusStep, types.DefaultNakamotoBonusPeriod)

	// highAvg = 20, lowAvg = 10, ratio = 2 < 3, should decrease
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(createValidators(20, 20, 10)).AnyTimes()

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, math.LegacyZeroDec(), nakamotoBonusCoefficient)
}

func TestAdjustEta_ClampZero(t *testing.T) {
	initEta := math.LegacyZeroDec()
	s := setupTestKeeper(t, initEta, types.DefaultNakamotoBonusPeriod)

	// highAvg = 20, lowAvg = 10, ratio = 2 < 3, should decrease, and clamp at 0
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(createValidators(20, 20, 10)).AnyTimes()

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.True(t, nakamotoBonusCoefficient.GTE(math.LegacyZeroDec()))
}

func TestAdjustEta_ClampOne(t *testing.T) {
	initEta := math.LegacyOneDec()
	s := setupTestKeeper(t, initEta, types.DefaultNakamotoBonusPeriod)

	// highAvg = 100, lowAvg = 10, ratio = 10 >= 3, should increase
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(createValidators(100, 100, 10)).AnyTimes()

	require.NoError(t, s.distrKeeper.AdjustEta(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.NakamotoBonus.Get(s.ctx)
	require.NoError(t, err)
	require.True(t, nakamotoBonusCoefficient.LTE(math.LegacyOneDec()))
}
