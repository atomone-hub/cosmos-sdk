package v5_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v5 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v5"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestMigrateStoreClearsDeprecatedGovParams(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(gov.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	govKey := storetypes.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)
	storeService := runtime.NewKVStoreService(govKey)
	sb := collections.NewSchemaBuilder(storeService)

	constitutionCollection := collections.NewItem(sb, v5.ConstitutionKey, "constitution", collections.StringValue)
	participationEMACollection := collections.NewItem(sb, v5.ParticipationEMAKey, "participation_ema", sdk.LegacyDecValue)
	constitutionAmendmentParticipationEMACollection := collections.NewItem(sb, v5.ConstitutionAmendmentParticipationEMAKey, "constitution_amendment_participation_ema", sdk.LegacyDecValue)
	lawParticipationEMACollection := collections.NewItem(sb, v5.LawParticipationEMAKey, "law_participation_ema", sdk.LegacyDecValue)

	legacyParams := legacyV4Params()
	store.Set(v5.ParamsKey, cdc.MustMarshal(&legacyParams))

	err := v5.MigrateStore(
		ctx,
		storeService,
		cdc,
		constitutionCollection,
		participationEMACollection,
		constitutionAmendmentParticipationEMACollection,
		lawParticipationEMACollection,
	)
	require.NoError(t, err)

	var params v1.Params
	err = cdc.Unmarshal(store.Get(v5.ParamsKey), &params)
	require.NoError(t, err)

	err = v1.ValidateGenesis(v1.NewGenesisState(
		v1.DefaultStartingProposalID,
		v1.DefaultParticipationEma,
		v1.DefaultParticipationEma,
		v1.DefaultParticipationEma,
		params,
	))
	require.NoErrorf(
		t,
		err,
		"migrated params still contain deprecated values: min_deposit=%v quorum=%q constitution_amendment_quorum=%q law_quorum=%q min_initial_deposit_ratio=%q",
		params.MinDeposit,
		params.Quorum,
		params.ConstitutionAmendmentQuorum,
		params.LawQuorum,
		params.MinInitialDepositRatio,
	)
}
