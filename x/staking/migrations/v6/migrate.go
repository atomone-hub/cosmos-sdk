package v6

import (
	corestoretypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateStore migrates the x/staking module state to version 6.
func MigrateStore(ctx sdk.Context, storeService corestoretypes.KVStoreService, cdc codec.BinaryCodec) error {
	// Open the KVStore
	store := storeService.OpenKVStore(ctx)

	paramsBz, err := store.Get(stakingtypes.ParamsKey)
	if err != nil {
		return err
	}

	var params stakingtypes.Params
	if err = cdc.Unmarshal(paramsBz, &params); err != nil {
		return err
	}

	defaultParams := stakingtypes.DefaultParams()
	params.MaxCommissionRate = defaultParams.MaxCommissionRate

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	return store.Set(stakingtypes.ParamsKey, bz)
}
