package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestBeginBlocker_NakamotoBonusEtaChange(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyNewDecWithPrec(3, 2), types.DefaultNakamotoBonusPeriod)

	// Create validators for mocking
	vals := createValidators(100, 100, 10)
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(vals, nil).AnyTimes()

	// Mock ValidatorByConsAddr for AllocateTokens
	for i := range vals {
		s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), gomock.Any()).Return(vals[i], nil).AnyTimes()
	}

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", types.ModuleName, fees)

	// Simulate BeginBlocker
	err := s.distrKeeper.BeginBlocker(s.ctx)
	require.NoError(t, err)

	// Verify η increased: 0.03 + 0.01 = 0.04
	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonus(s.ctx)
	require.NoError(t, err)

	expectedEta := math.LegacyNewDecWithPrec(4, 2)
	require.Equal(t, expectedEta, nakamotoBonusCoefficient,
		"η should increase from 0.03 to 0.04 when ratio >= 3. Got: %s", nakamotoBonusCoefficient)
}

func TestBeginBlocker_NakamotoBonusEtaDecrease(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyNewDecWithPrec(3, 2), types.DefaultNakamotoBonusPeriod)

	// Create validators with lower ratio for decrease
	vals := createValidators(20, 20, 10)
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(vals, nil).AnyTimes()

	// Mock ValidatorByConsAddr for AllocateTokens
	for i := range vals {
		s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), gomock.Any()).Return(vals[i], nil).AnyTimes()
	}

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", types.ModuleName, fees)

	// Simulate BeginBlocker
	err := s.distrKeeper.BeginBlocker(s.ctx)
	require.NoError(t, err)

	// Verify η clamped at minimum: 0.03 - 0.01 = 0.02, but min is 0.03, so stays at 0.03
	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonus(s.ctx)
	require.NoError(t, err)

	expectedEta := math.LegacyNewDecWithPrec(3, 2) // clamped at minimum 0.03
	require.Equal(t, expectedEta, nakamotoBonusCoefficient,
		"η should clamp at minimum (0.03) when decreasing below min. Got: %s", nakamotoBonusCoefficient)
}

func TestAllocateTokens_NakamotoBonusClampEta(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyOneDec(), types.DefaultNakamotoBonusPeriod)

	// η = 1.0, should clamp to 1.0 even if increase requested
	vals := createValidators(100, 100, 10)
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(vals, nil).AnyTimes()

	// Mock ValidatorByConsAddr for AllocateTokens
	for i := range vals {
		s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), gomock.Any()).Return(vals[i], nil).AnyTimes()
	}

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", types.ModuleName, fees)

	// Simulate BeginBlocker
	err := s.distrKeeper.BeginBlocker(s.ctx)
	require.NoError(t, err)

	// Should stay at 1 (clamped upper bound)
	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonus(s.ctx)
	require.NoError(t, err)
	require.Equal(t, math.LegacyOneDec(), nakamotoBonusCoefficient)
}

func TestAllocateTokens_NakamotoBonusClampEtaZero(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyZeroDec(), types.DefaultNakamotoBonusPeriod)

	// η = 0.0, should clamp to minimum (0.03) even though it's set to 0.0
	vals := createValidators(20, 20, 10)
	s.stakingKeeper.EXPECT().GetBondedValidatorsByPower(s.ctx).Return(vals, nil).AnyTimes()

	// Mock ValidatorByConsAddr for AllocateTokens
	for i := range vals {
		s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), gomock.Any()).Return(vals[i], nil).AnyTimes()
	}

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", types.ModuleName, fees)

	// Simulate BeginBlocker
	err := s.distrKeeper.BeginBlocker(s.ctx)
	require.NoError(t, err)

	// Should clamp to minimum (0.03) since 0.0 < min
	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonus(s.ctx)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDecWithPrec(3, 2), nakamotoBonusCoefficient,
		"η starting at 0 should clamp to minimum (0.03). Got: %s", nakamotoBonusCoefficient)
}
