package v1

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default period for deposits & voting
const (
	DefaultPeriod time.Duration = time.Hour * 24 * 2 // 2 days
)

// Default governance params
var (
	DefaultMinDepositTokens          = sdkmath.NewInt(10000000)
	DefaultQuorum                    = sdkmath.LegacyNewDecWithPrec(334, 3)
	DefaultThreshold                 = sdkmath.LegacyNewDecWithPrec(5, 1)
	DefaultMinInitialDepositRatio    = sdkmath.LegacyZeroDec()
	DefaultProposalCancelRatio       = sdkmath.LegacyMustNewDecFromStr("0.5")
	DefaultProposalCancelDestAddress = ""
	DefaultBurnProposalPrevote       = false // set to false to replicate behavior of when this change was made (0.47)
	DefaultBurnVoteQuorom            = false // set to false to  replicate behavior of when this change was made (0.47)
	DefaultMinDepositRatio           = sdkmath.LegacyMustNewDecFromStr("0.01")
)

// Deprecated: NewDepositParams creates a new DepositParams object
func NewDepositParams(minDeposit sdk.Coins, maxDepositPeriod *time.Duration) DepositParams {
	return DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: maxDepositPeriod,
	}
}

// Deprecated: NewTallyParams creates a new TallyParams object
func NewTallyParams(quorum, threshold string) TallyParams {
	return TallyParams{
		Quorum:    quorum,
		Threshold: threshold,
	}
}

// Deprecated: NewVotingParams creates a new VotingParams object
func NewVotingParams(votingPeriod *time.Duration) VotingParams {
	return VotingParams{
		VotingPeriod: votingPeriod,
	}
}

// NewParams creates a new Params instance with given values.
func NewParams(
	minDeposit sdk.Coins, maxDepositPeriod, votingPeriod time.Duration,
	quorum, threshold, minInitialDepositRatio, proposalCancelRatio, proposalCancelDest string,
	burnProposalDeposit, burnVoteQuorum bool, minDepositRatio string,
) Params {
	return Params{
		MinDeposit:                 minDeposit,
		MaxDepositPeriod:           &maxDepositPeriod,
		VotingPeriod:               &votingPeriod,
		Quorum:                     quorum,
		Threshold:                  threshold,
		MinInitialDepositRatio:     minInitialDepositRatio,
		ProposalCancelRatio:        proposalCancelRatio,
		ProposalCancelDest:         proposalCancelDest,
		BurnProposalDepositPrevote: burnProposalDeposit,
		BurnVoteQuorum:             burnVoteQuorum,
		MinDepositRatio:            minDepositRatio,
	}
}

// DefaultParams returns the default governance params
func DefaultParams() Params {
	return NewParams(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinDepositTokens)),
		DefaultPeriod,
		DefaultPeriod,
		DefaultQuorum.String(),
		DefaultThreshold.String(),
		DefaultMinInitialDepositRatio.String(),
		DefaultProposalCancelRatio.String(),
		DefaultProposalCancelDestAddress,
		DefaultBurnProposalPrevote,
		DefaultBurnVoteQuorom,
		DefaultMinDepositRatio.String(),
	)
}

// ValidateBasic performs basic validation on governance parameters.
func (p Params) ValidateBasic() error {
	minDeposit := sdk.Coins(p.MinDeposit)
	if minDeposit.Empty() || !minDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", minDeposit)
	}

	if p.MaxDepositPeriod == nil {
		return fmt.Errorf("maximum deposit period must not be nil: %d", p.MaxDepositPeriod)
	}

	if p.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", p.MaxDepositPeriod)
	}

	quorum, err := sdkmath.LegacyNewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("quorom too large: %s", p.Quorum)
	}

	threshold, err := sdkmath.LegacyNewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	if p.VotingPeriod == nil {
		return fmt.Errorf("voting period must not be nil: %d", p.VotingPeriod)
	}
	if p.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", p.VotingPeriod)
	}

	minInitialDepositRatio, err := sdkmath.LegacyNewDecFromStr(p.MinInitialDepositRatio)
	if err != nil {
		return fmt.Errorf("invalid mininum initial deposit ratio of proposal: %w", err)
	}
	if minInitialDepositRatio.IsNegative() {
		return fmt.Errorf("mininum initial deposit ratio of proposal must be positive: %s", minInitialDepositRatio)
	}
	if minInitialDepositRatio.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("mininum initial deposit ratio of proposal is too large: %s", minInitialDepositRatio)
	}

	proposalCancelRate, err := sdkmath.LegacyNewDecFromStr(p.ProposalCancelRatio)
	if err != nil {
		return fmt.Errorf("invalid burn rate of cancel proposal: %w", err)
	}
	if proposalCancelRate.IsNegative() {
		return fmt.Errorf("burn rate of cancel proposal must be positive: %s", proposalCancelRate)
	}
	if proposalCancelRate.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("burn rate of cancel proposal is too large: %s", proposalCancelRate)
	}

	if len(p.ProposalCancelDest) != 0 {
		_, err := sdk.AccAddressFromBech32(p.ProposalCancelDest)
		if err != nil {
			return fmt.Errorf("deposits destination address is invalid: %s", p.ProposalCancelDest)
		}
	}

	return nil
}
