package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SetGovernanceDelegation sets a governance delegation in the store
func (k Keeper) SetGovernanceDelegation(ctx sdk.Context, delegation v1.GovernanceDelegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&delegation)
	delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
	store.Set(types.GovernanceDelegationKey(delAddr), b)

	// Set the reverse mapping from governor to delegation
	// mainly for querying all delegations for a governor
	// TODO: see if we can avoid duplicate storage
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	store.Set(types.GovernanceDelegationsByGovernorKey(govAddr, delAddr), b)
}

// GetGovernanceDelegation gets a governance delegation from the store
func (k Keeper) GetGovernanceDelegation(ctx sdk.Context, delegatorAddr sdk.AccAddress) (v1.GovernanceDelegation, bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GovernanceDelegationKey(delegatorAddr))
	if b == nil {
		return v1.GovernanceDelegation{}, false
	}
	var delegation v1.GovernanceDelegation
	k.cdc.MustUnmarshal(b, &delegation)
	return delegation, true
}

// RemoveGovernanceDelegation removes a governance delegation from the store
func (k Keeper) RemoveGovernanceDelegation(ctx sdk.Context, delegatorAddr sdk.AccAddress) {
	// need to remove from both the delegator and governor mapping
	store := ctx.KVStore(k.storeKey)
	delegation, found := k.GetGovernanceDelegation(ctx, delegatorAddr)
	if !found {
		return
	}
	delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
	store.Delete(types.GovernanceDelegationKey(delAddr))

	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	store.Delete(types.GovernanceDelegationsByGovernorKey(govAddr, delAddr))
}

// IterateGovernorDelegations iterates over all governor delegations
func (k Keeper) IterateGovernorDelegations(ctx sdk.Context, governorAddr types.GovernorAddress, cb func(index int64, delegation v1.GovernanceDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GovernanceDelegationsByGovernorKey(governorAddr, []byte{}))
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		var delegation v1.GovernanceDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &delegation)
		if cb(i, delegation) {
			break
		}
		i++
	}
}

// GetAllGovernanceDelegationsByGovernor gets all governance delegations for a specific governor
func (k Keeper) GetAllGovernanceDelegationsByGovernor(ctx sdk.Context, governorAddr types.GovernorAddress) (delegations []*v1.GovernanceDelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GovernanceDelegationsByGovernorKey(governorAddr, []byte{}))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var delegation v1.GovernanceDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &delegation)
		delegations = append(delegations, &delegation)
	}
	return delegations
}

// IncreaseGovernorShares increases the governor validator shares in the store
func (k Keeper) IncreaseGovernorShares(ctx sdk.Context, governorAddr types.GovernorAddress, validatorAddr sdk.ValAddress, shares math.LegacyDec) {
	valShares, err := k.ValidatorSharesByGovernor.Get(ctx, collections.Join(governorAddr, validatorAddr))
	if err == collections.ErrEncoding {
		panic("error decoding governor validator shares")
	} else if err == collections.ErrNotFound {
		valShares = v1.NewGovernorValShares(governorAddr, validatorAddr, shares)
	} else {
		valShares.Shares = valShares.Shares.Add(shares)
	}
	k.ValidatorSharesByGovernor.Set(ctx, collections.Join(governorAddr, validatorAddr), valShares)
}

// DecreaseGovernorShares decreases the governor validator shares in the store
func (k Keeper) DecreaseGovernorShares(ctx sdk.Context, governorAddr types.GovernorAddress, validatorAddr sdk.ValAddress, shares math.LegacyDec) {
	share, err := k.ValidatorSharesByGovernor.Get(ctx, collections.Join(governorAddr, validatorAddr))
	if err == collections.ErrEncoding {
		panic("error decoding governor validator shares")
	} else if err == collections.ErrNotFound {
		panic("cannot decrease shares for a non-existent governor delegation")
	}
	share.Shares = share.Shares.Sub(shares)
	if share.Shares.IsNegative() {
		panic("negative shares")
	}
	if share.Shares.IsZero() {
		k.ValidatorSharesByGovernor.Remove(ctx, collections.Join(governorAddr, validatorAddr))
	} else {
		k.ValidatorSharesByGovernor.Set(ctx, collections.Join(governorAddr, validatorAddr), share)
	}
}

// UndelegateFromGovernor decreases all governor validator shares in the store
// and then removes the governor delegation for the given delegator
func (k Keeper) UndelegateFromGovernor(ctx sdk.Context, delegatorAddr sdk.AccAddress) error {
	delegation, found := k.GetGovernanceDelegation(ctx, delegatorAddr)
	if !found {
		return types.ErrGovernanceDelegationNotFound.Wrapf("governance delegation for delegator %s does not exist", delegatorAddr.String())
	}
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	// iterate all delegations of delegator and decrease shares
	err := k.sk.IterateDelegations(ctx, delegatorAddr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
		valAddr, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
		if err != nil {
			panic(err) // This should never happen
		}
		k.DecreaseGovernorShares(ctx, govAddr, valAddr, delegation.GetShares())
		return false
	})
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to iterate delegations: %v", err)
	}
	// remove the governor delegation
	k.RemoveGovernanceDelegation(ctx, delegatorAddr)
	return nil
}

// DelegateGovernor creates a governor delegation for the given delegator
// and increases all governor validator shares in the store
func (k Keeper) DelegateToGovernor(ctx sdk.Context, delegatorAddr sdk.AccAddress, governorAddr types.GovernorAddress) error {
	delegation := v1.NewGovernanceDelegation(delegatorAddr, governorAddr)
	k.SetGovernanceDelegation(ctx, delegation)
	// iterate all delegations of delegator and increase shares
	err := k.sk.IterateDelegations(ctx, delegatorAddr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
		valAddr, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
		if err != nil {
			panic(err) // This should never happen
		}
		k.IncreaseGovernorShares(ctx, governorAddr, valAddr, delegation.GetShares())
		return false
	})
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to iterate delegations: %v", err)
	}
	return nil
}

// RedelegateGovernor re-delegates all governor validator shares from one governor to another
func (k Keeper) RedelegateToGovernor(ctx sdk.Context, delegatorAddr sdk.AccAddress, dstGovernorAddr types.GovernorAddress) error {
	// undelegate from the source governor
	if err := k.UndelegateFromGovernor(ctx, delegatorAddr); err != nil {
		return err
	}
	// delegate to the destination governor
	return k.DelegateToGovernor(ctx, delegatorAddr, dstGovernorAddr)
}
