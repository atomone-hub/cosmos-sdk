package v5_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
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

func TestMigrateStore(t *testing.T) {
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

	var params v1.Params
	bz := store.Get(v5.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.Equal(t, legacyParams.MinDeposit, params.MinDeposit)
	require.Equal(t, legacyParams.Quorum, params.Quorum)
	require.Equal(t, legacyParams.ConstitutionAmendmentQuorum, params.ConstitutionAmendmentQuorum)
	require.Equal(t, legacyParams.LawQuorum, params.LawQuorum)
	require.Equal(t, legacyParams.MinInitialDepositRatio, params.MinInitialDepositRatio)

	// Run migrations.
	err := v5.MigrateStore(ctx, storeService, cdc, constitutionCollection, participationEMACollection, constitutionAmendmentParticipationEMACollection, lawParticipationEMACollection)
	require.NoError(t, err)

	// Check params
	bz = store.Get(v5.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.Empty(t, params.MinDeposit)
	require.Empty(t, params.Quorum)
	require.Empty(t, params.ConstitutionAmendmentQuorum)
	require.Empty(t, params.LawQuorum)
	require.Empty(t, params.MinInitialDepositRatio)
	require.Equal(t, v1.DefaultParams().MinDepositRatio, params.MinDepositRatio)
	require.NoError(t, params.ValidateBasic())
	require.NoError(t, v1.ValidateGenesis(v1.NewGenesisState(
		v1.DefaultStartingProposalID,
		v1.DefaultParticipationEma,
		v1.DefaultParticipationEma,
		v1.DefaultParticipationEma,
		params,
	)))

	// Check constitution
	result, err := constitutionCollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "This chain has no constitution.", result)

	// Check participation EMA values
	expectedEMA := math.LegacyNewDecWithPrec(12, 2)
	participationEMA, err := participationEMACollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedEMA, participationEMA)

	constitutionAmendmentEMA, err := constitutionAmendmentParticipationEMACollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedEMA, constitutionAmendmentEMA)

	lawEMA, err := lawParticipationEMACollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedEMA, lawEMA)
}

func legacyV4Params() v1.Params {
	maxDepositPeriod := v1.DefaultDepositPeriod
	votingPeriod := v1.DefaultVotingPeriod

	return v1.Params{
		MinDeposit:                     sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10_000_000))),
		MaxDepositPeriod:               &maxDepositPeriod,
		VotingPeriod:                   &votingPeriod,
		Quorum:                         "0.334000000000000000",
		Threshold:                      v1.DefaultThreshold.String(),
		MinInitialDepositRatio:         "0.100000000000000000",
		BurnVoteQuorum:                 v1.DefaultBurnVoteQuorom,
		BurnProposalDepositPrevote:     v1.DefaultBurnProposalPrevote,
		ConstitutionAmendmentQuorum:    v1.DefaultMinConstitutionAmendmentQuorum.String(),
		ConstitutionAmendmentThreshold: v1.DefaultConstitutionAmendmentThreshold.String(),
		LawQuorum:                      v1.DefaultMinLawQuorum.String(),
		LawThreshold:                   v1.DefaultLawThreshold.String(),
		QuorumTimeout:                  durationPtr(v1.DefaultQuorumTimeout),
		MaxVotingPeriodExtension:       durationPtr(v1.DefaultMaxVotingPeriodExtension),
	}
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
