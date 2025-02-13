import { ibcDenom } from "stridejs";
import { Chain } from "./types";

export const STRIDE_RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";
export const GAIA_RPC_ENDPOINT = "http://cosmoshub-rpc.internal.stridenet.co";
export const OSMO_RPC_ENDPOINT = "http://osmosis-rpc.internal.stridenet.co";

export const USTRD = "ustrd";
export const UATOM = "uatom";
export const UOSMO = "uosmo";

export const TRANSFER_CHANNEL: Record<Chain, Partial<Record<Chain, string>>> = {
  STRIDE: { GAIA: "channel-0", OSMO: "channel-1" },
  GAIA: { STRIDE: "channel-0" },
  OSMO: { STRIDE: "channel-0" },
};

export const ATOM_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: "transfer",
      incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
    },
  ],
  UATOM,
);

export const ATOM_DENOM_ON_OSMOSIS = ibcDenom(
  [
    {
      incomingPortId: "transfer",
      incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
    },
    {
      incomingPortId: "transfer",
      incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
    },
  ],
  UATOM,
);

export const STRD_DENOM_ON_OSMOSIS = ibcDenom(
  [
    {
      incomingPortId: "transfer",
      incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
    },
  ],
  USTRD,
);

export const STRD_DENOM_ON_GAIA = ibcDenom(
  [
    {
      incomingPortId: "transfer",
      incomingChannelId: TRANSFER_CHANNEL["GAIA"]["STRIDE"]!,
    },
  ],
  USTRD,
);

export const OSMO_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: "transfer",
      incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["OSMO"]!,
    },
  ],
  UOSMO,
);
