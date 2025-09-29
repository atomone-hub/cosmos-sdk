package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestGetMinDeposit(t *testing.T) {
	var (
		minDepositFloor   = v1.GetDefaultMinDepositFloor()
		minDepositFloorX2 = minDepositFloor.MulInt(math.NewInt(2))
		updatePeriod      = v1.DefaultMinDepositUpdatePeriod
		// N                 = v1.DefaultTargetActiveProposals // TODO

		// Handy function used to compute the min deposit time according to the
		// number of ticksPassed required.
		minDepositTimeFromTicks = func(ticks int) *time.Time {
			t := time.Now().Add(-updatePeriod*time.Duration(ticks) - time.Minute)
			return &t
		}
	)
	tests := []struct {
		name               string
		setup              func(sdk.Context, *keeper.Keeper)
		expectedMinDeposit string
	}{
		{
			name:               "initial case no setup : expectedMinDeposit=minDepositFloor",
			expectedMinDeposit: minDepositFloor.String(),
		},

		{
			name: "n=N-1 lastMinDeposit=minDepositFloor ticksPassed=0 : expectedMinDeposit=minDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloor,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: minDepositFloor.String(),
		},
		{
			name: "n=N lastMinDeposit=minDepositFloor ticksPassed=0 : expectedMinDeposit>minDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloor,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: "10500000stake",
		},
		{
			name: "n=N+1 lastMinDeposit=minDepositFloor ticksPassed=0 : expectedMinDeposit>minDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloor,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: "10500000stake",
		},
		{
			name: "n=N+1 lastMinDeposit=otherCoins ticksPassed=0 : expectedMinDeposit>minDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: sdk.NewCoins(
						sdk.NewInt64Coin("xxx", 1_000_000_000),
					),
					Time: minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: "10500000stake",
		},
		{
			name: "n=N-1 lastMinDeposit=minDepositFloor*2 ticksPassed=0 : minDeposit=minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloor,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: minDepositFloorX2.String(),
		},
		{
			name: "n=N lastMinDeposit=minDepositFloor*2 ticksPassed=0 : expectedMinDeposit>minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: "21000000stake",
		},
		{
			name: "n=N+1 lastMinDeposit=minDepositFloor*2 ticksPassed=0 : expectedMinDeposit>minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, false)
			},
			expectedMinDeposit: "21000000stake",
		},
		{
			name: "n=N+1 lastMinDeposit=minDepositFloor*2 ticksPassed=0 (try time-based update) : expectedMinDeposit=minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: minDepositFloorX2.String(),
		},
		{
			name: "n=N-1 lastMinDeposit=minDepositFloor*2 ticksPassed=1 : expectedMinDeposit<minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(1),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: "19500000stake",
		},
		{
			name: "n=N lastMinDeposit=minDepositFloor*2 ticksPassed=1 : expectedMinDeposit=minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(1),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: minDepositFloorX2.String(),
		},
		{
			name: "n=N+1 lastMinDeposit=minDepositFloor*2 ticksPassed=1 : expectedMinDeposit=minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(1),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: minDepositFloorX2.String(),
		},
		{
			name: "n=N-1 lastMinDeposit=minDepositFloor*2 ticksPassed=2 : expectedMinDeposit<minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(2),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx.WithBlockTime(*minDepositTimeFromTicks(1)), true)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: "19012500stake",
		},
		{
			name: "n=N lastMinDeposit=minDepositFloor*2 ticksPassed=2 : expectedMinDeposit=minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(2),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx.WithBlockTime(*minDepositTimeFromTicks(1)), true)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: minDepositFloorX2.String(),
		},
		{
			name: "n=N+1 lastMinDeposit=minDepositFloor*2 ticksPassed=2 : expectedMinDeposit=minDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minDepositFloorX2,
					Time:  minDepositTimeFromTicks(2),
				})
				require.NoError(t, err)
				k.UpdateMinDeposit(ctx.WithBlockTime(*minDepositTimeFromTicks(1)), true)
				k.UpdateMinDeposit(ctx, true)
			},
			expectedMinDeposit: minDepositFloorX2.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, _, _, _, _, _, ctx := setupGovKeeper(t)
			if tt.setup != nil {
				tt.setup(ctx, k)
			}

			minDeposit := k.GetMinDeposit(ctx)

			assert.Equal(t, tt.expectedMinDeposit, minDeposit.String())
		})
	}
}
