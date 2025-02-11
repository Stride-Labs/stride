import { StrideClient } from "stridejs";
import { coinFromString } from "stridejs";
import {
  cosmosProtoRegistry,
  ibcProtoRegistry,
  osmosis,
  osmosisProtoRegistry,
} from "osmojs";

const defaultTransferTimeout = BigInt(
  `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`,
);
import { MsgTransfer } from "stridejs/types/codegen/ibc/applications/transfer/v1/tx";
import { MsgRegisterTokenPriceQuery } from "stridejs/types/codegen/stride/icqoracle/tx";

/**
 * Creates a new transfer message for IBC transactions
 * @param stridejs The Stride client instance
 * @param channelId The transfer channel ID
 * @param coins The amount and denomination of coins to transfer as a string (e.g. "1000ustrd")
 * @param sender The address of the sender
 * @param receiver The address of the receiver
 * @param timeout Optional timeout for the IBC transfer, defaults to 3 minutes
 * @param memo Optional memo message for the transfer, defaults to empty string
 * @returns The IBC transfer message
 */
export function newTransferMsg({
  stridejs,
  channelId,
  coins,
  sender,
  receiver,
  memo = "",
  timeout,
}: {
  stridejs: StrideClient;
  channelId: string;
  coins: string;
  sender: string;
  receiver: string;
  timeout?: BigInt;
  memo?: string;
}): {
  typeUrl: string;
  value: MsgTransfer;
} {
  timeout = timeout === undefined ? timeout : defaultTransferTimeout;
  return stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
    {
      sourcePort: "transfer",
      sourceChannel: channelId,
      token: coinFromString(coins),
      sender: sender,
      receiver: receiver,
      timeoutHeight: {
        revisionNumber: 0n,
        revisionHeight: 0n,
      },
      timeoutTimestamp: defaultTransferTimeout,
      memo: memo,
    },
  );
}

export function newRegisterTokenPriceQueryMsg({
  adminClient,
  baseDenom,
  quoteDenom,
  baseDenomOnOsmosis,
  quoteDenomOnOsmosis,
  poolId,
  baseDenomDecimals = 6n,
  quoteDenomDecimals = 6n,
}: {
  adminClient: StrideClient;
  baseDenom: string;
  quoteDenom: string;
  baseDenomOnOsmosis: string;
  quoteDenomOnOsmosis: string;
  poolId: bigint;
  baseDenomDecimals?: bigint;
  quoteDenomDecimals?: bigint;
}): {
  typeUrl: string;
  value: MsgRegisterTokenPriceQuery;
} {
  return adminClient.types.stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery(
    {
      admin: adminClient.address,
      baseDenom: baseDenom,
      quoteDenom: quoteDenom,
      baseDenomDecimals: baseDenomDecimals,
      quoteDenomDecimals: quoteDenomDecimals,
      osmosisPoolId: poolId,
      osmosisBaseDenom: baseDenomOnOsmosis,
      osmosisQuoteDenom: quoteDenomOnOsmosis,
    },
  );
}

/**
 * Creates a new Osmosis gamm (balancer) pool creation message
 * @param sender - The address of the account creating the pool
 * @param tokens - Array of token denominations to include in the pool (e.g. ["10ibc/ustrd", "10uosmo"])
 * @param weights - Array of corresponding weights for each token in the pool (e.g. [1, 1])
 * @returns The gamm pool creation message
 */
export function newGammPoolMsg({
  sender,
  tokens,
  weights,
}: {
  sender: string;
  tokens: string[];
  weights: number[];
}) {
  if (tokens.length !== weights.length) {
    throw new Error("tokens and weights arrays must have the same length");
  }

  const poolAssets = tokens.map((token, index) => ({
    token: coinFromString(token),
    weight: weights[index].toString(),
  }));

  return osmosis.gamm.poolmodels.balancer.v1beta1.MessageComposer.withTypeUrl.createBalancerPool(
    {
      sender: sender,
      poolAssets: poolAssets,
      futurePoolGovernor: "",
      poolParams: {
        swapFee: "0.001",
        exitFee: "0",
      },
    },
  );
}
