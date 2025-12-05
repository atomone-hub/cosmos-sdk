package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SetGovernanceDelegation sets a governance delegation in the store
func (k Keeper) SetGovernanceDelegation(ctx sdk.Context, delegation v1.GovernanceDelegation) {
	delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
	k.GovernanceDelegations.Set(ctx, delAddr, delegation)

	// Set the reverse mapping from governor to delegation
	// mainly for querying all delegations for a governor
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	k.GovernanceDelegationsByGovernor.Set(ctx, collections.Join(govAddr, delAddr), delegation)
}

// RemoveGovernanceDelegation removes a governance delegation from the store
func (k Keeper) RemoveGovernanceDelegation(ctx sdk.Context, delegatorAddr sdk.AccAddress) {
	// need to remove from both the delegator and governor mapping
	delegation, err := k.GovernanceDelegations.Get(ctx, delegatorAddr)
	if err != nil {
		return
	}
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	k.GovernanceDelegations.Remove(ctx, delegatorAddr)
	k.GovernanceDelegationsByGovernor.Remove(ctx, collections.Join(govAddr, delegatorAddr))
}

// IncreaseGovernorShares increases the governor validator shares in the store
func (k Keeper) IncreaseGovernorShares(ctx sdk.Context, governorAddr types.GovernorAddress, validatorAddr sdk.ValAddress, shares math.LegacyDec) {
	valShares, err := k.ValidatorSharesByGovernor.Get(ctx, collections.Join(governorAddr, validatorAddr))
	if errors.IsOf(err, collections.ErrEncoding) {
		panic("error decoding governor validator shares")
	} else if errors.IsOf(err, collections.ErrNotFound) {
		valShares = v1.NewGovernorValShares(governorAddr, validatorAddr, shares)
	} else {
		valShares.Shares = valShares.Shares.Add(shares)
	}
	k.ValidatorSharesByGovernor.Set(ctx, collections.Join(governorAddr, validatorAddr), valShares)
}

// DecreaseGovernorShares decreases the governor validator shares in the store
func (k Keeper) DecreaseGovernorShares(ctx sdk.Context, governorAddr types.GovernorAddress, validatorAddr sdk.ValAddress, shares math.LegacyDec) {
	share, err := k.ValidatorSharesByGovernor.Get(ctx, collections.Join(governorAddr, validatorAddr))
	if errors.IsOf(err, collections.ErrEncoding) {
		panic("error decoding governor validator shares")
	} else if errors.IsOf(err, collections.ErrNotFound) {
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
	delegation, err := k.GovernanceDelegations.Get(ctx, delegatorAddr)
	if errors.IsOf(err, collections.ErrEncoding) {
		return sdkerrors.ErrInvalidRequest.Wrapf("error decoding governance delegation for delegator %s", delegatorAddr.String())
	} else if errors.IsOf(err, collections.ErrNotFound) {
		return types.ErrGovernanceDelegationNotFound.Wrapf("governance delegation for delegator %s does not exist", delegatorAddr.String())
	}
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	// iterate all delegations of delegator and decrease shares
	err = k.sk.IterateDelegations(ctx, delegatorAddr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
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
