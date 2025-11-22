package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var (
	StoreKey             = upgradetypes.StoreKey
	KeyUpgradedClient    = upgradetypes.KeyUpgradedClient
	KeyUpgradedConsState = upgradetypes.KeyUpgradedConsState
	KeyUpgradedIBCState  = upgradetypes.KeyUpgradedIBCState
)

func (p Plan) ValidateBasic() error {
	if !p.Time.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("time-based upgrades have been deprecated in the SDK")
	}
	if p.UpgradedClientState != nil {
		return sdkerrors.ErrInvalidRequest.Wrap("upgrade logic for IBC has been moved to the IBC module")
	}
	if len(p.Name) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if p.Height <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "height must be greater than 0")
	}

	return nil
}
