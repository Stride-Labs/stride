import { Secp256k1HdWallet } from "@cosmjs/amino";
import { GasPrice } from "@cosmjs/stargate";
import {
  StrideClient,
  ibcDenom,
  coinFromString,
  convertBech32Prefix,
} from "stridejs";
import { beforeAll, describe, test } from "vitest";
import { waitForChain, submitTxAndExpectSuccess } from "./utils";

const RPC_ENDPOINT = "https://stride-rpc.internal.stridenet.co";

let accounts: {
  user: StrideClient; // a normal account loaded with 100 STRD
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

  console.log("waiting for stride to start...");
  await waitForChain(accounts.user, "ustrd");
});

describe("x/stakeibc", () => {
  test("Registration", async () => {
    const stridejs = accounts.admin;

    console.log({
      creator: stridejs.address,
      bech32prefix: "cosmos",
      hostDenom: "uatom",
      ibcDenom: ibcDenom([{incomingPortId: "transfer", incomingChannelId: "channel-0"}], "uatom"),
      connectionId: "connection-0",
      transferChannelId: "channel-0",
      unbondingPeriod: BigInt(1),
      minRedemptionRate: "0.9",
      maxRedemptionRate: "1.5",
      lsmLiquidStakeEnabled: true,
      communityPoolTreasuryAddress: "",
      maxMessagesPerIcaTx: BigInt(2),
    })

    const msg = stridejs.types.stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone({
        creator: stridejs.address,
        bech32prefix: "cosmos",
        hostDenom: "uatom",
        ibcDenom: ibcDenom([{incomingPortId: "transfer", incomingChannelId: "channel-0"}], "uatom"),
        connectionId: "connection-0",
        transferChannelId: "channel-0",
        unbondingPeriod: BigInt(1),
        minRedemptionRate: "0.9",
        maxRedemptionRate: "1.5",
        lsmLiquidStakeEnabled: true,
        communityPoolTreasuryAddress: "",
        maxMessagesPerIcaTx: BigInt(2),
    });

    await submitTxAndExpectSuccess(stridejs, [msg]);
    console.log(stridejs.query.stride.stakeibc.hostZoneAll());
  });
});


