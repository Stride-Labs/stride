import { OfflineAminoSigner, Secp256k1HdWallet } from "@cosmjs/amino";
import { GasPrice, SigningStargateClient } from "@cosmjs/stargate";
import { fromSeconds } from "@cosmjs/tendermint-rpc";
import {
  coinFromString,
  convertBech32Prefix,
  decToString,
  StrideClient,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import { waitForChain } from "./utils";

const STRIDE_RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";
const GAIA_RPC_ENDPOINT = "http://cosmoshub-rpc.internal.stridenet.co";
const OSMO_RPC_ENDPOINT = "http://osmosis-rpc.internal.stridenet.co";

const TRANSFER_CHANNEL = {
  STRIDE: { GAIA: "channel-0", OSMO: "channel-1" },
  GAIA: { STRIDE: "channel-0" },
  OSMO: { STRIDE: "channel-0" },
};

let accounts: {
  user: StrideClient; // a normal account loaded with 100 STRD
  admin: StrideClient; // the stride admin account loaded with 1000 STRD
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
};

export type GaiaClient = {
  signer: OfflineAminoSigner;
  client: SigningStargateClient;
};

export function isGaiaClient(client: any): client is GaiaClient {
  return (
    "signer" in client &&
    "getAccounts" in client.signer &&
    "signAmino" in client.signer &&
    "client" in client &&
    client.client instanceof SigningStargateClient
  );
}

let gaiaAccounts: {
  user: GaiaClient; // a normal account loaded with 100 ATOM
};

const mnemonics: {
  name: "user" | "admin" | "val1" | "val2" | "val3";
  mnemonic: string;
}[] = [
  {
    name: "user",
    mnemonic:
      "brief play describe burden half aim soccer carbon hope wait output play vacuum joke energy crucial output mimic cruise brother document rail anger leaf",
  },
  {
    name: "admin",
    mnemonic:
      "tone cause tribe this switch near host damage idle fragile antique tail soda alien depth write wool they rapid unfold body scan pledge soft",
  },
  {
    name: "val1",
    mnemonic:
      "close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly",
  },
  {
    name: "val2",
    mnemonic:
      "turkey miss hurry unable embark hospital kangaroo nuclear outside term toy fall buffalo book opinion such moral meadow wing olive camp sad metal banner",
  },
  {
    name: "val3",
    mnemonic:
      "tenant neck ask season exist hill churn rice convince shock modify evidence armor track army street stay light program harvest now settle feed wheat",
  },
];

// init accounts and wait for chain to start
beforeAll(async () => {
  console.log("setting up accounts...");
  // @ts-expect-error
  // init accounts as an empty object, then add the accounts in the loop
  accounts = {};
  for (const { name, mnemonic } of mnemonics) {
    // setup signer
    //
    // IMPORTANT: We're using Secp256k1HdWallet from @cosmjs/amino because sending amino txs tests both amino and direct.
    // That's because the tx contains the direct encoding anyway, and also attaches a signature on the amino encoding.
    // The mempool then converts from direct to amino to verify the signature.
    // Therefore if the signature verification passes, we can be sure that both amino and direct are working properly.
    const signer = await Secp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: "stride",
    });

    // get signer address
    const [{ address }] = await signer.getAccounts();

    accounts[name] = await StrideClient.create(
      STRIDE_RPC_ENDPOINT,
      signer,
      address,
      {
        gasPrice: GasPrice.fromString("0.025ustrd"),
        broadcastPollIntervalMs: 50,
        resolveIbcResponsesCheckIntervalMs: 50,
      },
    );

    if (name === "user") {
      const signer = await Secp256k1HdWallet.fromMnemonic(mnemonic);

      // get signer address
      const [{ address }] = await signer.getAccounts();

      gaiaAccounts = {
        user: {
          signer,
          client: await SigningStargateClient.connectWithSigner(
            GAIA_RPC_ENDPOINT,
            signer,
            {
              gasPrice: GasPrice.fromString("1.0uatom"),
              broadcastPollIntervalMs: 50,
            },
          ),
        },
      };

      // TODO osmosisAccount
    }
  }
  console.log("waiting for stride to start...");
  await waitForChain(accounts.user, "ustrd");

  console.log("waiting for gaia to start...");
  await waitForChain(gaiaAccounts.user, "uatom");
});

// time variables in seconds
const now = () => Math.floor(Date.now() / 1000);
const minute = 60;
const hour = 60 * minute;
const day = 24 * hour;

describe("x/airdrop", () => {
  test("MsgCreateAirdrop", async () => {
    const stridejs = accounts.admin;

    const nowSec = now();
    const airdropId = String(nowSec);

    const msg =
      stridejs.types.stride.airdrop.MessageComposer.withTypeUrl.createAirdrop({
        admin: stridejs.address,
        airdropId: airdropId,
        rewardDenom: "ustrd",
        distributionStartDate: fromSeconds(now()),
        distributionEndDate: fromSeconds(nowSec + 3 * day),
        clawbackDate: fromSeconds(nowSec + 4 * day),
        claimTypeDeadlineDate: fromSeconds(nowSec + 2 * day),
        earlyClaimPenalty: decToString(0.5),
        allocatorAddress: stridejs.address,
        distributorAddress: stridejs.address,
        linkerAddress: stridejs.address,
      });

    const tx = await stridejs.signAndBroadcast([msg], 2);

    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const { id, earlyClaimPenalty } =
      await stridejs.query.stride.airdrop.airdrop({
        id: airdropId,
      });

    expect(id).toBe(airdropId);
    expect(earlyClaimPenalty).toBe("0.5");
  });
});

describe("ibc", () => {
  test("MsgTransfer", async () => {
    const stridejs = accounts.user;

    const msg =
      stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
        {
          sourcePort: "transfer",
          sourceChannel: TRANSFER_CHANNEL["STRIDE"]["GAIA"],
          token: coinFromString("1ustrd"),
          sender: stridejs.address,
          receiver: convertBech32Prefix(stridejs.address, "cosmos"),
          timeoutHeight: {
            revisionNumber: 0n,
            revisionHeight: 0n,
          },
          timeoutTimestamp: BigInt(
            `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes
          ),
          memo: "",
        },
      );

    const tx = await stridejs.signAndBroadcast([msg], 2);
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const ibcAck = await tx.ibcResponses[0];
    expect(ibcAck.type).toBe("ack");
    expect(ibcAck.tx.code).toBe(0);
  }, 30_000);
});

describe("x/icqoracle", () => {
  test.only("happy path", async () => {
    // - Transfer STRD to Osmosis
    // - Transfer ATOM to Osmosis
    // - Create STRD/OSMO pool
    // - Create ATOM/OSMO pool
    // - Add TokenPrice(base=STRD, quote=OSMO)
    // - Add TokenPrice(base=ATOM, quote=OSMO)
    // - Query for price of ATOM in STRD

    // Transfer STRD to Osmosis
    const stridejs = accounts.user;

    const strideTx = await stridejs.signAndBroadcast([
      stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
        {
          sourcePort: "transfer",
          sourceChannel: TRANSFER_CHANNEL["STRIDE"]["OSMO"],
          token: coinFromString("1000000ustrd"),
          sender: stridejs.address,
          receiver: convertBech32Prefix(stridejs.address, "osmo"),
          timeoutHeight: {
            revisionNumber: 0n,
            revisionHeight: 0n,
          },
          timeoutTimestamp: BigInt(
            `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes
          ),
          memo: "",
        },
      ),
    ]);
    if (strideTx.code !== 0) {
      console.error(strideTx.rawLog);
    }
    expect(strideTx.code).toBe(0);

    const ibcAck = await strideTx.ibcResponses[0];
    expect(ibcAck.type).toBe("ack");
    expect(ibcAck.tx.code).toBe(0);

    // Transfer ATOM to Osmosis
    const gaiajs = gaiaAccounts.user;

    const [{ address: gaiaAddress }] = await gaiajs.signer.getAccounts();

    const gaiaTx = await gaiajs.client.signAndBroadcast(
      gaiaAddress,
      [
        {
          typeUrl: "/ibc.applications.transfer.v1.MsgTransfer",
          value: {
            sourcePort: "transfer",
            sourceChannel: TRANSFER_CHANNEL["GAIA"]["STRIDE"],
            token: coinFromString("1000000uatom"),
            sender: gaiaAddress,
            receiver: convertBech32Prefix(gaiaAddress, "stride"), // needs to be valid but ignored by pfm
            timeoutHeight: {
              revisionNumber: 0n,
              revisionHeight: 0n,
            },
            timeoutTimestamp: BigInt(
              `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes
            ),
            memo: JSON.stringify({
              forward: {
                receiver: convertBech32Prefix(gaiaAddress, "osmo"),
                port: "transfer",
                channel: TRANSFER_CHANNEL["STRIDE"]["OSMO"],
              },
            }),
          },
        },
      ],
      "auto",
    );

    if (gaiaTx.code !== 0) {
      console.error(gaiaTx.rawLog);
    }
    expect(gaiaTx.code).toBe(0);

    // // packet forward should resolve only after the final destination is acked
    // ibcAck = await tx.ibcResponses[0];
    // expect(ibcAck.type).toBe("ack");
    // expect(ibcAck.tx.code).toBe(0);
  }, 120_000);
});
