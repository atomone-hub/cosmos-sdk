package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (k Keeper) getGovernorBondedTokens(ctx sdk.Context, govAddr types.GovernorAddress) (bondedTokens math.Int, err error) {
	bondedTokens = math.ZeroInt()
	addr := sdk.AccAddress(govAddr)
	err = k.sk.IterateDelegations(ctx, addr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
		if err != nil {
			panic(err) // This should never happen
		}
		validator, _ := k.sk.GetValidator(ctx, validatorAddr)
		shares := delegation.GetShares()
		bt := shares.MulInt(validator.GetBondedTokens()).Quo(validator.GetDelegatorShares()).TruncateInt()
		bondedTokens = bondedTokens.Add(bt)

		return false
	})
	if err != nil {
		return math.ZeroInt(), errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "failed to iterate delegations: %v", err)
	}

	return bondedTokens, nil
}

func (k Keeper) ValidateGovernorMinSelfDelegation(ctx sdk.Context, governor v1.Governor) bool {
	// ensure that the governor is active and that has a valid governance self-delegation
	if !governor.IsActive() {
		return false
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to get gov params: %v", err))
	}
	minGovernorSelfDelegation, _ := math.NewIntFromString(params.MinGovernorSelfDelegation)
	bondedTokens, err := k.getGovernorBondedTokens(ctx, governor.GetAddress())
	if err != nil {
		return false
	}
	delAddr := sdk.AccAddress(governor.GetAddress())

	if del, found := k.GetGovernanceDelegation(ctx, delAddr); !found || governor.GovernorAddress != del.GovernorAddress {
		panic("active governor without governance self-delegation")
	}

	if bondedTokens.LT(minGovernorSelfDelegation) {
		return false
	}

	return true
}
