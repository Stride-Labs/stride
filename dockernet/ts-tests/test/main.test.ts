import { Secp256k1HdWallet } from "@cosmjs/amino";
import { Registry } from "@cosmjs/proto-signing";
import {
  AminoTypes,
  defaultRegistryTypes,
  SigningStargateClient,
} from "@cosmjs/stargate";
import {
  cosmos,
  cosmosAminoConverters,
  getSigningStrideClient,
  stride,
  strideAminoConverters,
  strideProtoRegistry,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import { feeFromGas, sleep } from "./utils";
import { fromRfc3339WithNanoseconds } from "@cosmjs/tendermint-rpc";

const RPC_ENDPOINT = "http://localhost:26657";

type Account = {
  signer: Awaited<ReturnType<typeof Secp256k1HdWallet.fromMnemonic>>;
  address: string;
  query: Awaited<ReturnType<typeof stride.ClientFactory.createRPCQueryClient>>;
  tx: Awaited<ReturnType<typeof getSigningStrideClient>>;
};

let accounts: {
  user: Account; // just a normal user account loaded with 100 STRD
  admin: Account; // the stride admin account loaded with 1000 STRD
  val1: Account;
  val2: Account;
  val3: Account;
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
    // @ts-expect-error
    // init accounts[name] as an empty object, then add the fields one by one
    accounts[name] = {};

    // setup signer
    //
    // IMPORTANT: we're using Secp256k1HdWallet from @cosmjs/amino because sending amino txs tests both amino and direcy.
    // that's because the tx contains the direct encoding anyway, and also attaches a signature on the amino encoding.
    // the mempool then converts from direct to amino to verify the signature.
    // therefore if the signature verification passes, we can be sure that both amino and direct are supported.
    accounts[name].signer = await Secp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: "stride",
    });

    // setup address
    const [{ address }] = await accounts[name].signer.getAccounts();
    accounts[name].address = address;

    // setup query client
    accounts[name].query = await stride.ClientFactory.createRPCQueryClient({
      rpcEndpoint: RPC_ENDPOINT,
    });

    // setup tx client
    const registry = new Registry([
      ...defaultRegistryTypes,
      ...strideProtoRegistry,
    ]);
    const aminoTypes = new AminoTypes({
      ...strideAminoConverters,
      ...cosmosAminoConverters,
    });

    accounts[name].tx = await SigningStargateClient.connectWithSigner(
      RPC_ENDPOINT,
      accounts[name].signer,
      {
        registry,
        aminoTypes,
      },
    );
  }

  console.log("waiting for chain to start...");
  while (true) {
    const block =
      await accounts.user.query.cosmos.base.tendermint.v1beta1.getLatestBlock(
        {},
      );

    if (block.block.header.height.toNumber() > 0) {
      break;
    }

    await sleep(50);
  }
});

describe("x/airdrop", () => {
  test("create airdrop", async () => {
    const msg = stride.airdrop.MessageComposer.withTypeUrl.createAirdrop({
      admin: accounts.admin.address,
      airdropId: "🍌",
      rewardDenom: "ustrd",
      distributionStartDate: fromRfc3339WithNanoseconds(
        new Date().toISOString(),
      ),
      distributionEndDate: fromRfc3339WithNanoseconds(new Date().toISOString()),
      clawbackDate: fromRfc3339WithNanoseconds(new Date().toISOString()),
      claimTypeDeadlineDate: fromRfc3339WithNanoseconds(
        new Date().toISOString(),
      ),
      earlyClaimPenalty: "5",
      distributionAddress: accounts.val1.address,
    });

    const tx = await accounts.user.tx.signAndBroadcast(
      accounts.user.address,
      [msg],
      feeFromGas(200000),
    );

    expect(tx.code).toBe(0);
  });
});
