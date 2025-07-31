import { ibcDenom } from "stridejs";
import { Chain, ChainConfig, ChainConfigs } from "./types";

export const STRIDE_RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";
export const GAIA_RPC_ENDPOINT = "http://cosmoshub-rpc.internal.stridenet.co";
export const OSMO_RPC_ENDPOINT = "http://osmosis-rpc.internal.stridenet.co";

export const STRIDE_CHAIN_ID = "stride-test-1";
export const GAIA_CHAIN_ID = "cosmoshub-test-1";
export const OSMO_CHAIN_ID = "osmosis-test-1";

export const USTRD = "ustrd";
export const UATOM = "uatom";
export const UOSMO = "uosmo";

export const toStToken = (denom: string) => `st${denom}`;
export const STATOM = toStToken(UATOM);
export const STOSMO = toStToken(UOSMO);

export const DEFAULT_FEE = BigInt(2000000);
export const DEFAULT_GAS = "2000000";

export const TRANSFER_PORT = "transfer";

export const REMOVED = "REMOVED";

export const DEFAULT_CONNECTION_ID = "connection-0";
export const DEFAULT_TRANSFER_CHANNEL_ID = "channel-0";

export const STRIDE_CHAIN_NAME = "stride";
export const COSMOSHUB_CHAIN_NAME = "cosmoshub";
export const OSMOSIS_CHAIN_NAME = "osmosis";

// NOTE: This assumes only one host zone is up at a time
export function newChainConfig({
  chainId,
  hostDenom,
  bechPrefix,
  coinType,
  connectionId,
  transferChannelId,
  rpcEndpoint,
}: {
  chainId: string;
  hostDenom: string;
  bechPrefix: string;
  coinType: number;
  connectionId: string;
  transferChannelId: string;
  rpcEndpoint: string;
}): ChainConfig {
  return {
    chainId,
    hostDenom,
    bechPrefix,
    coinType,
    connectionId,
    transferChannelId,
    rpcEndpoint,
    stDenom: toStToken(hostDenom),
    strdDenomOnHost: ibcDenom(
      [
        {
          incomingPortId: TRANSFER_PORT,
          incomingChannelId: transferChannelId,
        },
      ],
      USTRD,
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
  };
}

export const CHAIN_CONFIGS: ChainConfigs = {
  cosmoshub: newChainConfig({
    chainId: GAIA_CHAIN_ID,
    hostDenom: UATOM,
    bechPrefix: "cosmos",
    coinType: 118,
    connectionId: DEFAULT_CONNECTION_ID,
    transferChannelId: DEFAULT_TRANSFER_CHANNEL_ID,
    rpcEndpoint: GAIA_RPC_ENDPOINT,
  }),
  osmosis: newChainConfig({
    chainId: OSMO_CHAIN_ID,
    hostDenom: UOSMO,
    bechPrefix: "osmo",
    coinType: 118,
    connectionId: DEFAULT_CONNECTION_ID,
    transferChannelId: DEFAULT_TRANSFER_CHANNEL_ID,
    rpcEndpoint: OSMO_RPC_ENDPOINT,
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
