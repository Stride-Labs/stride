import { osmosis } from "osmojs";
import { Coin, coinFromString, ibc, stride } from "stridejs";
import { MsgTransfer } from "stridejs/dist/types/codegen/ibc/applications/transfer/v1/tx";
import { MsgRegisterTokenPriceQuery } from "stridejs/dist/types/codegen/stride/icqoracle/tx";
import { UOSMO } from "./consts";

function TransferTimeoutSec(sec: number) {
  return BigInt(`${Math.floor(Date.now() / 1000) + sec}000000000`);
}

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
  channelId,
  coin,
  sender,
  receiver,
  memo = "",
  timeout,
}: {
  channelId: string;
  coin: string;
  sender: string;
  receiver: string;
  timeout?: BigInt;
  memo?: string;
}): {
  typeUrl: string;
  value: MsgTransfer;
} {
  timeout = timeout === undefined ? timeout : TransferTimeoutSec(60);
  return ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer({
    sourcePort: "transfer",
    sourceChannel: channelId,
    token: coinFromString(coin),
    sender: sender,
    receiver: receiver,
    timeoutHeight: {
      revisionNumber: 0n,
      revisionHeight: 0n,
    },
    timeoutTimestamp: TransferTimeoutSec(60),
    memo: memo,
  });
}

export function newRegisterTokenPriceQueryMsg({
  admin,
  baseDenom,
  quoteDenom,
  baseDenomOnOsmosis,
  quoteDenomOnOsmosis,
  poolId,
}: {
  admin: string;
  baseDenom: string;
  quoteDenom: string;
  baseDenomOnOsmosis: string;
  quoteDenomOnOsmosis: string;
  poolId: bigint;
}): {
  typeUrl: string;
  value: MsgRegisterTokenPriceQuery;
} {
  return stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
    admin: admin,
    baseDenom: baseDenom,
    quoteDenom: quoteDenom,
    osmosisPoolId: poolId,
    osmosisBaseDenom: baseDenomOnOsmosis,
    osmosisQuoteDenom: quoteDenomOnOsmosis,
  });
}

/**
 * Creates a new Osmosis gamm (balancer) pool creation message
 * @param sender - The address of the account creating the pool
 * @param tokens - Array of token denominations to include in the pool (e.g. ["10ibc/ustrd", "10uosmo"])
 * @param weights - Array of corresponding weights for each token in the pool (e.g. [1, 1])
 * @returns The gamm pool creation message
 */
export function newGammPoolMsg({ sender, tokens, weights }: { sender: string; tokens: string[]; weights: number[] }) {
  if (tokens.length !== weights.length) {
    throw new Error("tokens and weights arrays must have the same length");
  }

  const poolAssets = tokens.map((token, index) => ({
    token: coinFromString(token),
    weight: weights[index].toString(),
  }));

  return osmosis.gamm.poolmodels.balancer.v1beta1.MessageComposer.withTypeUrl.createBalancerPool({
    sender: sender,
    poolAssets: poolAssets,
    futurePoolGovernor: "",
    poolParams: {
      swapFee: "0.001",
      exitFee: "0",
    },
  });
}

/**
 * denom1 is always "uosmo"
 */
export function newConcentratedLiquidityPoolMsg({ sender, denom0 }: { sender: string; denom0: string }) {
  return osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.MessageComposer.withTypeUrl.createConcentratedPool(
    {
      sender,
      denom0,
      denom1: UOSMO,
      tickSpacing: 100n,
      spreadFactor: "0.001",
    },
  );
}

/**
 * since denom1 is always "uosmo", `tokenMinAmount1` should be denominated in `uosmo` and `tokensProvided` should have uosmos at index 1
 */
export function addConcentratedLiquidityPositionMsg({
  sender,
  poolId,
  tokensProvided,
  tokenMinAmount0,
  tokenMinAmount1,
}: {
  sender: string;
  poolId: bigint;
  tokensProvided: Coin[];
  tokenMinAmount0: string;
  tokenMinAmount1: string;
}) {
  if (tokensProvided.length !== 2) {
    throw new Error("tokensProvided.length must be 2");
  }
  if (BigInt(tokenMinAmount0) > BigInt(tokensProvided[0].amount)) {
    throw new Error("tokenMinAmount0 bigger than provided");
  }
  if (BigInt(tokenMinAmount1) > BigInt(tokensProvided[1].amount)) {
    throw new Error("tokenMinAmount1 bigger than provided");
  }

  return osmosis.concentratedliquidity.v1beta1.MessageComposer.withTypeUrl.createPosition({
    sender,
    poolId,
    lowerTick: -108000000n,
    upperTick: 342000000n,
    tokensProvided,
    tokenMinAmount0,
    tokenMinAmount1,
  });
}
