package v4

import (
	"fmt"

	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dstrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// MigrateStore migrates the x/distribution module state to version 6.
func MigrateStore(ctx sdk.Context, storeService corestoretypes.KVStoreService, cdc codec.BinaryCodec) error {
	// Open the KVStore
	store := storeService.OpenKVStore(ctx)

	paramsBz, err := store.Get(dstrtypes.ParamsKey)
	if err != nil {
		return err
	}

	var params dstrtypes.Params
	if err = cdc.Unmarshal(paramsBz, &params); err != nil {
		return err
	}

	defaultParams := dstrtypes.DefaultParams()
	params.NakamotoBonus = defaultParams.NakamotoBonus

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	if err := store.Set(dstrtypes.ParamsKey, bz); err != nil {
		return err
	}

	// Check if NakamotoBonus parameter already exists
	nakamotoBonusKey := dstrtypes.NakamotoBonusKey
	exists, err := store.Has(nakamotoBonusKey)
	if err != nil {
		return fmt.Errorf("error checking if nakamoto bonus key exists: %w", err)
	}
	if !exists {
		// Set the default value
		// defaultNakamotoBonus := dstrtypes.DefaultNakamotoBonus
		// Marshal and set the parameter in the store
		//if err := store.Set(nakamotoBonusKey, cdc.MustMarshal(&defaultNakamotoBonus)); err != nil {
		//	return err
		//}
	}

	return nil
}
