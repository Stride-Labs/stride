import { ibcDenom } from "stridejs";
import { Chain, ChainConfig, ChainConfigs, Mnemonics } from "./types";
import keysData from "../../../network/configs/keys.json";

export const MNEMONICS = keysData as Mnemonics;

export const STRIDE_CHAIN_NAME = "stride";
export const COSMOSHUB_CHAIN_NAME = "cosmoshub";
export const OSMOSIS_CHAIN_NAME = "osmosis";

export const STRIDE_RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";

export const USTRD = "ustrd";
export const STRD = "ustrd";
export const UATOM = "uatom";
export const UOSMO = "uosmo";

export const DEFAULT_FEE = BigInt(2000000);
export const DEFAULT_GAS = "2000000";

export const TRANSFER_PORT = "transfer";

export const REMOVED = "REMOVED";

export const DEFAULT_CONNECTION_ID = "connection-0";
export const DEFAULT_TRANSFER_CHANNEL_ID = "channel-0";

// NOTE: This assumes only one host zone is up at a time
export function newChainConfig({
  chainName,
  hostDenom,
  bechPrefix,
  coinType,
  connectionId,
  transferChannelId,
  rpcEndpoint,
}: {
  chainName: string;
  hostDenom: string;
  bechPrefix: string;
  coinType: number;
  connectionId: string;
  transferChannelId: string;
  rpcEndpoint: string;
}): ChainConfig {
  return {
    chainName,
    chainId: `${chainName}-test-1`,
    hostDenom,
    bechPrefix,
    coinType,
    connectionId,
    transferChannelId,
    rpcEndpoint,
    stDenom: `st${hostDenom}`,
    strdDenomOnHost: ibcDenom(
      [
        {
          incomingPortId: TRANSFER_PORT,
          incomingChannelId: DEFAULT_TRANSFER_CHANNEL_ID, // assumes each host only has 1 ibc connection (to stride)
        },
      ],
      STRD,
    ),
    hostDenomOnStride: ibcDenom(
      [
        {
          incomingPortId: TRANSFER_PORT,
          incomingChannelId: transferChannelId,
        },
      ],
      hostDenom,
    ),
    stDenomOnHost: ibcDenom(
      [
        {
          incomingPortId: TRANSFER_PORT,
          incomingChannelId: DEFAULT_TRANSFER_CHANNEL_ID, // assumes each host only has 1 ibc connection (to stride)
        },
      ],
      `st${hostDenom}`,
    ),
  };
}

export const TEST_CHAINS = ["cosmoshub", "osmosis"];

export const CHAIN_CONFIGS: ChainConfigs = {
  cosmoshub: newChainConfig({
    chainName: "cosmoshub",
    hostDenom: "uatom",
    bechPrefix: "cosmos",
    coinType: 118,
    connectionId: "connection-0",
    transferChannelId: "channel-0",
    rpcEndpoint: "http://cosmoshub-rpc.internal.stridenet.co",
  }),
  osmosis: newChainConfig({
    chainName: "osmosis",
    hostDenom: "uosmo",
    bechPrefix: "osmo",
    coinType: 118,
    connectionId: "connection-1",
    transferChannelId: "channel-1",
    rpcEndpoint: "http://osmosis-rpc.internal.stridenet.co",
  }),
};

export const TRANSFER_CHANNEL: Record<Chain, Partial<Record<Chain, string>>> = {
  stride: { cosmoshub: "channel-0", osmosis: "channel-1" },
  cosmoshub: { stride: "channel-0" },
  osmosis: { stride: "channel-0" },
};

export const CONNECTION_ID: Record<Chain, Partial<Record<Chain, string>>> = {
  stride: { osmosis: "connection-0" },
  cosmoshub: { stride: "connection-0" },
  osmosis: { stride: "connection-0" },
};

export const ATOM_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][COSMOSHUB_CHAIN_NAME]!,
    },
  ],
  UATOM,
);

export const ATOM_DENOM_ON_OSMOSIS = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][COSMOSHUB_CHAIN_NAME]!,
    },
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL[OSMOSIS_CHAIN_NAME][STRIDE_CHAIN_NAME]!,
    },
  ],
  UATOM,
);

export const STRD_DENOM_ON_OSMOSIS = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL[OSMOSIS_CHAIN_NAME][STRIDE_CHAIN_NAME]!,
    },
  ],
  USTRD,
);

export const STRD_DENOM_ON_GAIA = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL[COSMOSHUB_CHAIN_NAME][STRIDE_CHAIN_NAME]!,
    },
  ],
  USTRD,
);

export const OSMO_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][OSMOSIS_CHAIN_NAME]!,
    },
  ],
  UOSMO,
);
