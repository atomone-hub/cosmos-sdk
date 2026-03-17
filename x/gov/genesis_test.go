package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestImportExportQueues_ErrorUnconsistentState(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	require.Panics(t, func() {
		gov.InitGenesis(ctx, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, &v1.GenesisState{
			Deposits: v1.Deposits{
				{
					ProposalId: 1234,
					Depositor:  "me",
					Amount: sdk.Coins{
						sdk.NewCoin(
							"stake",
							sdkmath.NewInt(1234),
						),
					},
				},
			},
		})
	})
	gov.InitGenesis(ctx, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, v1.DefaultGenesisState())
	genState, err := gov.ExportGenesis(ctx, suite.GovKeeper)
	require.NoError(t, err)

	// Compare core fields (LastMinDeposit and LastMinInitialDeposit are dynamic and set during InitGenesis)
	expected := v1.DefaultGenesisState()
	require.Equal(t, expected.StartingProposalId, genState.StartingProposalId)
	require.Equal(t, expected.Params, genState.Params)
	require.Equal(t, expected.Constitution, genState.Constitution)
	require.Equal(t, expected.ParticipationEma, genState.ParticipationEma)
	require.Equal(t, expected.ConstitutionAmendmentParticipationEma, genState.ConstitutionAmendmentParticipationEma)
	require.Equal(t, expected.LawParticipationEma, genState.LawParticipationEma)
	require.Empty(t, genState.Deposits)
	require.Empty(t, genState.Votes)
	require.Empty(t, genState.Proposals)

	// Verify that dynamic deposit values were initialized
	require.NotNil(t, genState.LastMinDeposit)
	require.NotNil(t, genState.LastMinInitialDeposit)
	require.NotEmpty(t, genState.LastMinDeposit.Value)
	require.NotEmpty(t, genState.LastMinInitialDeposit.Value)
}

// TestInitGenesis_NonGovernorDelegation verifies that a non-governor account
// can delegate to a governor.
func TestInitGenesis_NonGovernorDelegation(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)

	// Create two accounts: one will be a governor, the other a plain delegator
	governorPubKey := pubkeys[0]
	delegatorPubKey := pubkeys[1]

	governorAccAddr := sdk.AccAddress(governorPubKey.Address())
	delegatorAccAddr := sdk.AccAddress(delegatorPubKey.Address())

	// Register both accounts in the auth module
	govAcc := suite.AccountKeeper.NewAccountWithAddress(ctx, governorAccAddr)
	suite.AccountKeeper.SetAccount(ctx, govAcc)

	delAcc := suite.AccountKeeper.NewAccountWithAddress(ctx, delegatorAccAddr)
	suite.AccountKeeper.SetAccount(ctx, delAcc)

	govAddr := types.GovernorAddress(governorAccAddr)
	now := time.Now().UTC()

	governor, err := v1.NewGovernor(govAddr.String(), v1.NewGovernorDescription("test-gov", "", "", "", ""), now)
	require.NoError(t, err)

	defaultState := v1.DefaultGenesisState()
	defaultState.Governors = []*v1.Governor{&governor}
	defaultState.GovernanceDelegations = []*v1.GovernanceDelegation{
		{
			DelegatorAddress: delegatorAccAddr.String(),
			GovernorAddress:  govAddr.String(),
		},
	}

	require.NotPanics(t, func() {
		gov.InitGenesis(ctx, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, defaultState)
	})
}
