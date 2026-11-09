package solanablockrewards

import (
	"fmt"
	"strconv"

	"github.com/rpcpool/yellowstone-faithful/jsonbuilder"
	"github.com/rpcpool/yellowstone-faithful/third_party/solana_proto/confirmed_block"
)

func RewardsToUi(
	rewards *confirmed_block.Rewards,
) (*jsonbuilder.ArrayBuilder, *uint64, error) {
	rewardsArray := jsonbuilder.NewArray()

	for _, reward := range rewards.Rewards {
		rewardJson := jsonbuilder.NewObject()
		{
			rewardJson.String("pubkey", reward.Pubkey)
			rewardJson.Int("lamports", reward.Lamports)
			rewardJson.Uint("postBalance", reward.PostBalance)
			if rewardType, ok := RewardTypeToUi(reward.RewardType); ok {
				rewardJson.String("rewardType", rewardType)
			} else {
				rewardJson.Null("rewardType")
			}
			if reward.Commission != "" {
				rewardJson.Float("commission", asFloat(reward.Commission))
			} else {
				rewardJson.Null("commission")
			}
			// commissionBps (SIMD-0291) is skipped rather than nulled when absent,
			// matching #[serde(skip_serializing_if = "Option::is_none")] on
			// Reward::commission_bps in agave. A value that does not parse is
			// skipped too, so it does not read as a real 0% commission.
			if reward.CommissionBps != "" {
				if commissionBps, err := strconv.ParseUint(reward.CommissionBps, 10, 16); err == nil {
					rewardJson.Uint("commissionBps", commissionBps)
				}
			}
		}
		rewardsArray.AddObject(rewardJson)
	}
	if rewards.NumPartitions != nil {
		numPart := rewards.NumPartitions.NumPartitions
		return rewardsArray, &numPart, nil
	}
	return rewardsArray, nil, nil
}

// RewardTypeToUi returns the JSON name of a reward type (Agave's serde variant
// name). Unspecified and unknown values have none and serialize as null, as in Agave.
func RewardTypeToUi(t confirmed_block.RewardType) (string, bool) {
	if t == confirmed_block.RewardType_Unspecified {
		return "", false
	}
	name, ok := confirmed_block.RewardType_name[int32(t)]
	return name, ok
}

func asFloat(s string) float64 {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0
	}
	return f
}
