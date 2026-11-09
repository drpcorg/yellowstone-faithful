package solanablockrewards

import (
	"testing"

	"github.com/rpcpool/yellowstone-faithful/third_party/solana_proto/confirmed_block"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRewardsToUi(t *testing.T) {
	rewards := &confirmed_block.Rewards{
		Rewards: []*confirmed_block.Reward{
			{Pubkey: "a", Lamports: 1, PostBalance: 2, RewardType: confirmed_block.RewardType_Fee},
			{Pubkey: "b", Lamports: -3, PostBalance: 4, RewardType: confirmed_block.RewardType_DeactivatedStake},
			{Pubkey: "c", Lamports: -5, PostBalance: 6, RewardType: confirmed_block.RewardType_VATDebit},
			{Pubkey: "d", Lamports: 7, PostBalance: 8, RewardType: confirmed_block.RewardType_Voting, Commission: "10"},
			{Pubkey: "e", Lamports: 9, PostBalance: 10, RewardType: confirmed_block.RewardType(99)},
		},
		NumPartitions: &confirmed_block.NumPartitions{NumPartitions: 43},
	}
	arr, numPartitions, err := RewardsToUi(rewards)
	require.NoError(t, err)
	defer arr.Put()
	require.NotNil(t, numPartitions)
	require.Equal(t, uint64(43), *numPartitions)

	got, err := arr.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `[
		{"pubkey":"a","lamports":1,"postBalance":2,"rewardType":"Fee","commission":null},
		{"pubkey":"b","lamports":-3,"postBalance":4,"rewardType":"DeactivatedStake","commission":null},
		{"pubkey":"c","lamports":-5,"postBalance":6,"rewardType":"VATDebit","commission":null},
		{"pubkey":"d","lamports":7,"postBalance":8,"rewardType":"Voting","commission":10},
		{"pubkey":"e","lamports":9,"postBalance":10,"rewardType":null,"commission":null}
	]`, string(got))

	_, numPartitions, err = RewardsToUi(&confirmed_block.Rewards{})
	require.NoError(t, err)
	require.Nil(t, numPartitions)
}

// Field-for-field parity with what agave 4.2 returns for the same rewards,
// round-tripped through the protobuf schema. The first reward is the one from
// slot 440208047 that surfaced the missing commissionBps.
func TestRewardsToUiAgaveParity(t *testing.T) {
	buf, err := proto.Marshal(&confirmed_block.Rewards{
		Rewards: []*confirmed_block.Reward{
			{
				Pubkey:        "3ZFkP99s4GvhibGHUtQnF14bnZfmpDmbDd8xeUV37mC8",
				Lamports:      81902744,
				PostBalance:   279084185624,
				RewardType:    confirmed_block.RewardType_DeactivatedStake,
				CommissionBps: "0",
			},
			{
				Pubkey:        "GdnSyH3YtwcxFvQrVVJMm1JhTS4QVX7MFsX56uJLUfiZ",
				Lamports:      2500,
				PostBalance:   100000,
				RewardType:    confirmed_block.RewardType_Voting,
				Commission:    "7",
				CommissionBps: "725",
			},
			{
				Pubkey:      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
				Lamports:    -5000,
				PostBalance: 42,
				RewardType:  confirmed_block.RewardType_Fee,
			},
			{
				Pubkey:        "11111111111111111111111111111111",
				Lamports:      1,
				PostBalance:   2,
				RewardType:    confirmed_block.RewardType_Staking,
				CommissionBps: "not a number",
			},
		},
	})
	require.NoError(t, err)

	rewards, err := ParseRewards(buf)
	require.NoError(t, err)

	arr, numPartitions, err := RewardsToUi(rewards)
	require.NoError(t, err)
	defer arr.Put()
	require.Nil(t, numPartitions)

	got, err := arr.MarshalJSON()
	require.NoError(t, err)
	// Exact bytes, not JSONEq: field order is part of the parity with agave.
	const want = `[` +
		`{"pubkey":"3ZFkP99s4GvhibGHUtQnF14bnZfmpDmbDd8xeUV37mC8","lamports":81902744,"postBalance":279084185624,"rewardType":"DeactivatedStake","commission":null,"commissionBps":0},` +
		`{"pubkey":"GdnSyH3YtwcxFvQrVVJMm1JhTS4QVX7MFsX56uJLUfiZ","lamports":2500,"postBalance":100000,"rewardType":"Voting","commission":7,"commissionBps":725},` +
		`{"pubkey":"9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM","lamports":-5000,"postBalance":42,"rewardType":"Fee","commission":null},` +
		`{"pubkey":"11111111111111111111111111111111","lamports":1,"postBalance":2,"rewardType":"Staking","commission":null}` +
		`]`
	require.Equal(t, want, string(got))
}
