import { Secp256k1HdWallet } from "@cosmjs/amino";
import { GasPrice } from "@cosmjs/stargate";
import { fromSeconds } from "@cosmjs/tendermint-rpc";
import {
  coinFromString,
  convertBech32Prefix,
  decToString,
  sleep,
  StrideClient,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";

const RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";

let accounts: {
  user: StrideClient; // just a normal user account loaded with 100 STRD
  admin: StrideClient; // the stride admin account loaded with 1000 STRD
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
};

// init accounts and wait for chain to start
beforeAll(async () => {
  console.log("setting up accounts...");

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

  // @ts-expect-error
  // init accounts as an empty object, then add the accounts in the loop
  accounts = {};
  for (const { name, mnemonic } of mnemonics) {
    // setup signer
    //
    // IMPORTANT: we're using Secp256k1HdWallet from @cosmjs/amino because sending amino txs tests both amino and direct.
    // that's because the tx contains the direct encoding anyway, and also attaches a signature on the amino encoding.
    // the mempool then converts from direct to amino to verify the signature.
    // therefore if the signature verification passes, we can be sure that both amino and direct are supported.
    const signer = await Secp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: "stride",
    });

    // get signer address
    const [{ address }] = await signer.getAccounts();

    accounts[name] = await StrideClient.create(RPC_ENDPOINT, signer, address, {
      gasPrice: GasPrice.fromString("0.025ustrd"),
      broadcastPollIntervalMs: 50,
      resolveIbcResponsesCheckIntervalMs: 50,
    });
  }

  console.log("waiting for chain to start...");
  while (true) {
    try {
      const block =
        await accounts.user.query.cosmos.base.tendermint.v1beta1.getLatestBlock(
          {},
        );

      if (block?.block?.header.height! > 0n) {
        break;
      }

      await sleep(50);
    } catch (error) {}
  }
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

    const { airdrop } = await stridejs.query.stride.airdrop.airdrop({
      id: airdropId,
    });

    expect(airdrop!.id).toBe(airdropId);
    expect(airdrop!.earlyClaimPenalty).toBe("0.5");
  });
});

describe("ibc", () => {
  test("MsgTransfer", async () => {
    const stridejs = accounts.user;

    const msg =
      stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
        {
          sourcePort: "transfer",
          sourceChannel: "channel-0",
          token: coinFromString("1ustrd"),
          sender: stridejs.address,
          receiver: convertBech32Prefix(stridejs.address, "cosmos"),
          timeoutHeight: {
            revisionNumber: 0n,
            revisionHeight: 0n,
          },
          timeoutTimestamp: BigInt(
            `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes from now as nanoseconds
          ),
          memo: "",
        },
      );

    const tx = await stridejs.signAndBroadcast([msg]);
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const ibcAck = await tx.ibcResponses[0];
    expect(ibcAck.type).toBe("ack");
    expect(ibcAck.tx.code).toBe(0);
  }, 30_000);
});
