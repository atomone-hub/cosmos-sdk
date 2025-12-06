package v4

import (
	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dstrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// MigrateStore migrates the x/distribution module state to version 6.
func MigrateStore(
	ctx sdk.Context,
	storeService corestoretypes.KVStoreService,
	cdc codec.BinaryCodec,
	nakamotoBonus collections.Item[math.LegacyDec],
) error {
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

	defaultNakamotoBonus := dstrtypes.DefaultNakamotoBonus
	if ok, err := nakamotoBonus.Has(ctx); !ok || err != nil {
		if err := nakamotoBonus.Set(ctx, defaultNakamotoBonus); err != nil {
			return err
		}
	}

	return nil
}
