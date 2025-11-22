package types

import (
	"context"

	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func NewUpgradeKeeperCompat(k *upgradekeeper.Keeper) *upgradeKeeperCompat {
	return &upgradeKeeperCompat{
		Keeper: k,
	}
}

type upgradeKeeperCompat struct {
	*upgradekeeper.Keeper
}

func (u *upgradeKeeperCompat) GetUpgradePlan(ctx context.Context) (Plan, error) {
	plan, err := u.Keeper.GetUpgradePlan(ctx)
	if err != nil {
		return Plan{}, err
	}

	return Plan{
		Name:   plan.Name,
		Height: plan.Height,
		Info:   plan.Info,
	}, nil
}

func (u *upgradeKeeperCompat) GetUpgradedClient(ctx context.Context, height int64) ([]byte, error) {
	return u.Keeper.GetUpgradedClient(ctx, height)
}

func (u *upgradeKeeperCompat) GetUpgradedConsensusState(ctx context.Context, lastHeight int64) ([]byte, error) {
	return u.Keeper.GetUpgradedConsensusState(ctx, lastHeight)
}

func (u *upgradeKeeperCompat) ScheduleUpgrade(ctx context.Context, plan Plan) error {
	return u.Keeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   plan.Name,
		Height: plan.Height,
		Info:   plan.Info,
	})
}

func (u *upgradeKeeperCompat) SetUpgradedClient(ctx context.Context, planHeight int64, bz []byte) error {
	return u.Keeper.SetUpgradedClient(ctx, planHeight, bz)
}

func (u *upgradeKeeperCompat) SetUpgradedConsensusState(ctx context.Context, planHeight int64, bz []byte) error {
	return u.Keeper.SetUpgradedConsensusState(ctx, planHeight, bz)
}
