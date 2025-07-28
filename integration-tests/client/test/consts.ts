import { ibcDenom } from "stridejs";
import { Chain } from "./types";

export const STRIDE_RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";
export const GAIA_RPC_ENDPOINT = "http://cosmoshub-rpc.internal.stridenet.co";
export const OSMO_RPC_ENDPOINT = "http://osmosis-rpc.internal.stridenet.co";

export const STRIDE_CHAIN_ID = "stride-test-1";
export const GAIA_CHAIN_ID = "cosmoshub-test-1";
export const OSMO_CHAIN_ID = "osmosis-test-1";

export const USTRD = "ustrd";
export const UATOM = "uatom";
export const UOSMO = "uosmo";

const toStToken = (denom: string) => `st${denom}`;
export const STATOM = toStToken(UATOM);
export const STOSMO = toStToken(UOSMO);

export const DEFAULT_FEE = 2000000;
export const DEFAULT_GAS = "2000000";

export const TRANSFER_PORT = "transfer";

export const TRANSFER_CHANNEL: Record<Chain, Partial<Record<Chain, string>>> = {
  STRIDE: { GAIA: "channel-0", OSMO: "channel-1" },
  GAIA: { STRIDE: "channel-0" },
  OSMO: { STRIDE: "channel-0" },
};

export const CONNECTION_ID: Record<Chain, Partial<Record<Chain, string>>> = {
  STRIDE: { GAIA: "connection-0", OSMO: "connection-1" },
  GAIA: { STRIDE: "connection-0" },
  OSMO: { STRIDE: "connection-0" },
};

export const ATOM_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
    },
  ],
  UATOM,
);

export const ATOM_DENOM_ON_OSMOSIS = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
    },
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
    },
  ],
  UATOM,
);

export const STRD_DENOM_ON_OSMOSIS = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
    },
  ],
  USTRD,
);

export const STRD_DENOM_ON_GAIA = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL["GAIA"]["STRIDE"]!,
    },
  ],
  USTRD,
);

export const OSMO_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["OSMO"]!,
    },
  ],
  UOSMO,
);
