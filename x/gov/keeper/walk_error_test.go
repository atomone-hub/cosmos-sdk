package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestUndelegateGovernor_ReturnsWalkError(t *testing.T) {
	govKeeper, storeKey, _, _, _, _, _, ctx := setupGovKeeperWithStoreKey(t)

	addrs := simtestutil.CreateRandomAccounts(3)
	delegatorAddr := addrs[0]
	corruptDelegatorAddr := addrs[1]
	governorAddr := types.GovernorAddress(addrs[2])

	governor, err := v1.NewGovernor(governorAddr.String(), v1.GovernorDescription{}, time.Now())
	require.NoError(t, err)
	governor.Status = v1.Inactive
	require.NoError(t, govKeeper.Governors.Set(ctx, governor.GetAddress(), governor))

	govKeeper.SetGovernanceDelegation(ctx, v1.NewGovernanceDelegation(delegatorAddr, governor.GetAddress()))

	corruptKey, err := collections.EncodeKeyWithPrefix(
		types.GovernanceDelegationsByGovernorKeyPrefix,
		govKeeper.GovernanceDelegationsByGovernor.KeyCodec(),
		collections.Join(governor.GetAddress(), corruptDelegatorAddr),
	)
	require.NoError(t, err)
	ctx.KVStore(storeKey).Set(corruptKey, []byte{0x01})

	cacheCtx, _ := ctx.CacheContext()
	msgServer := keeper.NewMsgServerImpl(govKeeper)

	_, err = msgServer.UndelegateGovernor(cacheCtx, v1.NewMsgUndelegateGovernor(delegatorAddr))
	require.Error(t, err)
}

func TestGovernorsDelegationsInvariant_ReturnsGovernorWalkError(t *testing.T) {
	govKeeper, storeKey, _, _, stakingKeeper, _, _, ctx := setupGovKeeperWithStoreKey(t)

	governorAddr := types.GovernorAddress(simtestutil.CreateRandomAccounts(1)[0])
	corruptKey, err := collections.EncodeKeyWithPrefix(
		types.GovernorsKeyPrefix,
		govKeeper.Governors.KeyCodec(),
		governorAddr,
	)
	require.NoError(t, err)
	ctx.KVStore(storeKey).Set(corruptKey, []byte{0x01})

	invariant := keeper.GovernorsDelegationsInvariant(govKeeper, stakingKeeper)
	msg, broken := invariant(ctx)

	require.True(t, broken)
	require.Contains(t, msg, "failed to iterate governors")
}

func TestGovernorsDelegationsInvariant_ReturnsValidatorSharesWalkError(t *testing.T) {
	govKeeper, storeKey, _, _, stakingKeeper, _, _, ctx := setupGovKeeperWithStoreKey(t)

	addrs := simtestutil.CreateRandomAccounts(2)
	governorAddr := types.GovernorAddress(addrs[0])
	validatorAddr := sdk.ValAddress(addrs[1])

	governor, err := v1.NewGovernor(governorAddr.String(), v1.GovernorDescription{}, time.Now())
	require.NoError(t, err)
	governor.Status = v1.Inactive
	require.NoError(t, govKeeper.Governors.Set(ctx, governor.GetAddress(), governor))

	corruptKey, err := collections.EncodeKeyWithPrefix(
		types.ValidatorSharesByGovernorKeyPrefix,
		govKeeper.ValidatorSharesByGovernor.KeyCodec(),
		collections.Join(governor.GetAddress(), validatorAddr),
	)
	require.NoError(t, err)
	ctx.KVStore(storeKey).Set(corruptKey, []byte{0x01})

	invariant := keeper.GovernorsDelegationsInvariant(govKeeper, stakingKeeper)
	msg, broken := invariant(ctx)

	require.True(t, broken)
	require.Contains(t, msg, "failed to iterate validator shares")
}
