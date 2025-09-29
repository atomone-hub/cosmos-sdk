package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// Simulation parameter constants
const (
	MinDeposit         = "min_deposit"
	DepositPeriod      = "deposit_period"
	MinInitialRatio    = "min_initial_ratio"
	VotingPeriod       = "voting_period"
	Quorum             = "quorum"
	Threshold          = "threshold"
	ProposalCancelRate = "proposal_cancel_rate"
	MinDepositRatio    = "min_deposit_ratio"

	// tallyMax must be at least as large as the regular Threshold
	// Therefore, we use this break out point in randomization.
	tallyMax = 500
)

// GenDepositPeriod returns randomized DepositPeriod
func GenDepositPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenMinDeposit returns randomized MinDeposit
func GenMinDeposit(r *rand.Rand, bondDenom string) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(bondDenom, int64(simulation.RandIntBetween(r, 1, 1e3/2))))
}

// GenDepositMinInitialRatio returns randomized DepositMinInitialRatio
func GenDepositMinInitialDepositRatio(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(sdkmath.LegacyNewDec(100))
}

// GenProposalCancelRate returns randomized ProposalCancelRate
func GenProposalCancelRate(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(sdkmath.LegacyNewDec(100))
}

// GenVotingPeriod returns randomized VotingPeriod
func GenVotingPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 0, 99)) * time.Second
}

// GenQuorum returns randomized Quorum
func GenQuorum(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 334, 500)), 3)
}

// GenThreshold returns randomized Threshold
func GenThreshold(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 450, tallyMax+1)), 3)
}

// GenVeto returns randomized Veto
func GenVeto(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 250, 334)), 3)
}

// GenMinDepositRatio returns randomized DepositMinRatio
func GenMinDepositRatio(*rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyMustNewDecFromStr("0.01")
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	startingProposalID := uint64(simState.Rand.Intn(100))

	var minDeposit sdk.Coins
	simState.AppParams.GetOrGenerate(MinDeposit, &minDeposit, simState.Rand, func(r *rand.Rand) { minDeposit = GenMinDeposit(r, simState.BondDenom) })

	var depositPeriod time.Duration
	simState.AppParams.GetOrGenerate(DepositPeriod, &depositPeriod, simState.Rand, func(r *rand.Rand) { depositPeriod = GenDepositPeriod(r) })

	var minInitialDepositRatio sdkmath.LegacyDec
	simState.AppParams.GetOrGenerate(MinInitialRatio, &minInitialDepositRatio, simState.Rand, func(r *rand.Rand) { minInitialDepositRatio = GenDepositMinInitialDepositRatio(r) })

	var proposalCancelRate sdkmath.LegacyDec
	simState.AppParams.GetOrGenerate(ProposalCancelRate, &proposalCancelRate, simState.Rand, func(r *rand.Rand) { proposalCancelRate = GenProposalCancelRate(r) })

	var votingPeriod time.Duration
	simState.AppParams.GetOrGenerate(VotingPeriod, &votingPeriod, simState.Rand, func(r *rand.Rand) { votingPeriod = GenVotingPeriod(r) })

	var quorum sdkmath.LegacyDec
	simState.AppParams.GetOrGenerate(Quorum, &quorum, simState.Rand, func(r *rand.Rand) { quorum = GenQuorum(r) })

	var threshold sdkmath.LegacyDec
	simState.AppParams.GetOrGenerate(Threshold, &threshold, simState.Rand, func(r *rand.Rand) { threshold = GenThreshold(r) })

	var minDepositRatio sdkmath.LegacyDec
	simState.AppParams.GetOrGenerate(MinDepositRatio, &minDepositRatio, simState.Rand, func(r *rand.Rand) { minDepositRatio = GenMinDepositRatio(r) })

	govGenesis := v1.NewGenesisState(
		startingProposalID,
		v1.NewParams(minDeposit, depositPeriod, votingPeriod, quorum.String(), threshold.String(), minInitialDepositRatio.String(), proposalCancelRate.String(), "", simState.Rand.Intn(2) == 0, simState.Rand.Intn(2) == 0, minDepositRatio.String()),
	)

	bz, err := json.MarshalIndent(&govGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(govGenesis)
}
