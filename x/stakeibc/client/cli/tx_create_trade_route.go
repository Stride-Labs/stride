package cli

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

type TradeRouteDef struct {
	HostChainId                  string `json:"host_chain_id,omitempty"`
	HostConnectionId             string `json:"host_connection_id,omitempty"`
	HostIcaAddress               string `json:"host_ica_address,omitempty"`
	RewardChainId                string `json:"reward_chain_id,omitempty"`
	RewardConnectionId           string `json:"reward_connection_id,omitempty"`
	RewardIcaAddress             string `json:"reward_ica_address,omitempty"`
	TradeChainId                 string `json:"trade_chain_id,omitempty"`
	TradeConnectionId            string `json:"trade_connection_id,omitempty"`
	TradeIcaAddress              string `json:"trade_ica_address,omitempty"`
	HostRewardTransferChannelId  string `json:"host_reward_transfer_channel_id,omitempty"`
	RewardTradeTransferChannelId string `json:"reward_trade_transfer_channel_id,omitempty"`
	TradeHostTransferChannelId   string `json:"trade_host_transfer_channel_id,omitempty"`
	RewardDenomOnHost            string `json:"reward_denom_on_host,omitempty"`
	RewardDenomOnReward          string `json:"reward_denom_on_reward,omitempty"`
	RewardDenomOnTrade           string `json:"reward_denom_on_trade,omitempty"`
	TargetDenomOnTrade           string `json:"target_denom_on_trade,omitempty"`
	TargetDenomOnHost            string `json:"target_denom_on_host,omitempty"`
	PoolId                       string `json:"pool_id,omitempty"`
	MinSwapAmount                string `json:"min_swap_amount,omitempty"`
	MaxSwapAmount                string `json:"max_swap_amount,omitempty"`
}

// Parse a JSON with fields for a new trade route in the format
//
//	{
//		  "host_chain_id": "hostChain-1",
//		  "host_connection_id": "17",
//		  "host_ica_address": "host8d0a8d663af02932",
//
//		  "reward_chain_id": "rewardChain-04",
//		  "reward_connection_id": "9",
//		  "reward_ica_address": "reward0028391a90fee",
//
//		  "trade_chain_id": "tradeChain-2",
//		  "trade_connection_id": "24",
//		  "trade_ica_address": "trade82a7cee9f0da12",
//
//		  "host_reward_transfer_channel_id": "3",
//		  "reward_trade_transfer_channel_id": "6",
//		  "trade_host_transfer_channel_id": "12",
//
//		  "reward_denom_on_host": "ibc/840982463847",
//		  "reward_denom_on_reward": "rewardToken",
//		  "reward_denom_on_trade": "ibc/02738646723",
//		  "target_denom_on_trade": "ibc/79227384721",
//		  "target_denom_on_host": "hostToken",
//
//		  "pool_id": "273",
//		  "min_swap_amount": "0",
//		  "max_swap_amount": "1000000000000",
//	}
//
// Notice that every field value is a string, even the last 3

func parseCreateTradeRouteFile(validatorsFile string) (newTradeRoute TradeRouteDef, err error) {
	fileContents, err := os.ReadFile(validatorsFile)
	if err != nil {
		return newTradeRoute, err
	}

	if err = json.Unmarshal(fileContents, &newTradeRoute); err != nil {
		return newTradeRoute, err
	}

	return newTradeRoute, nil
}

func CmdCreateTradeRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-trade-route [trade-route-file]",
		Short: "Broadcast message create-trade-route",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tradeRouteProposalFile := args[0]

			newRoute, err := parseCreateTradeRouteFile(tradeRouteProposalFile)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(newRoute.PoolId, 10, 64)
			minSwapAmount, found := sdk.NewIntFromString(newRoute.MinSwapAmount)
			if !found {
				minSwapAmount = sdk.ZeroInt()
			}
			maxSwapAmount, found := sdk.NewIntFromString(newRoute.MaxSwapAmount)
			if !found {
				const MaxUint = ^uint(0)
				const MaxInt = int64(MaxUint >> 1)
				maxSwapAmount = sdk.NewInt(MaxInt)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateTradeRoute(
				clientCtx.GetFromAddress().String(),
				newRoute.HostChainId,
				newRoute.HostConnectionId,
				newRoute.HostIcaAddress,
				newRoute.RewardChainId,
				newRoute.RewardConnectionId,
				newRoute.RewardIcaAddress,
				newRoute.TradeChainId,
				newRoute.TradeConnectionId,
				newRoute.TradeIcaAddress,
				newRoute.HostRewardTransferChannelId,
				newRoute.RewardTradeTransferChannelId,
				newRoute.TradeHostTransferChannelId,
				newRoute.RewardDenomOnHost,
				newRoute.RewardDenomOnReward,
				newRoute.RewardDenomOnTrade,
				newRoute.TargetDenomOnTrade,
				newRoute.TargetDenomOnHost,
				poolId,
				minSwapAmount,
				maxSwapAmount,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
