package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress        = "set_withdraw_address"
	EventTypeRewards                   = "rewards"
	EventTypeCommission                = "commission"
	EventTypeWithdrawRewards           = "withdraw_rewards"
	EventTypeWithdrawCommission        = "withdraw_commission"
	EventTypeProposerReward            = "proposer_reward"
	EventTypeUpdateNakamotoCoefficient = "update_nakamoto_coefficient"

	AttributeKeyWithdrawAddress  = "withdraw_address"
	AttributeKeyValidator        = "validator"
	AttributeKeyDelegator        = "delegator"
	AttributeNakamotoCoefficient = "nakamoto_coefficient"
)
