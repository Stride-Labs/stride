import { Secp256k1HdWallet } from "@cosmjs/amino";
import { getSigningStrideClient, stride, cosmos } from "stridejs";
import { beforeAll, expect, test } from "vitest";
import { coinsFromString } from "./utils";

const RPC_ENDPOINT = "http://localhost:26657";

type Account = {
  signer: Awaited<ReturnType<typeof Secp256k1HdWallet.fromMnemonic>>;
  address: string;
  query: Awaited<ReturnType<typeof stride.ClientFactory.createRPCQueryClient>>;
  tx: Awaited<ReturnType<typeof getSigningStrideClient>>;
};

let accounts: {
  user: Account;
  val1: Account;
  val2: Account;
  val3: Account;
};

// init accounts and wait for chain to start
beforeAll(async () => {
  const mnemonics: {
    name: "user" | "val1" | "val2" | "val3";
    mnemonic: string;
  }[] = [
    {
      name: "user",
      mnemonic:
        "brief play describe burden half aim soccer carbon hope wait output play vacuum joke energy crucial output mimic cruise brother document rail anger leaf",
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

  //@ts-expect-error initialize accounts as an empty object, then add the accounts in the loop
  accounts = {};
  for (const { name, mnemonic } of mnemonics) {
    //@ts-expect-error ts cries about accounts[name] not having any of the declared fields
    // which we're going to add a few lines down
    accounts[name] = {};

    // setup signer
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
    accounts[name].tx = await getSigningStrideClient({
      rpcEndpoint: RPC_ENDPOINT,
      signer: accounts[name].signer,
    });

    const balance = await accounts[name].query.cosmos.bank.v1beta1.allBalances({
      address: accounts[name].address,
    });
    console.log(
      "balance",
      name,
      BigInt(balance.balances[0].amount) / BigInt(1e6),
      balance.balances[0].denom,
    );
  }

  const msgSend = cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
    fromAddress: accounts.user.address,
    toAddress: accounts.user.address,
    amount: [{ amount: "1", denom: "ustrd" }],
  });

  const tx = await accounts.user.tx.signAndBroadcast(
    accounts.user.address,
    [msgSend],
    { amount: coinsFromString("0.025ustrd"), gas: "200000" },
  );

  console.log(tx.code);
});

test("adds 1 + 2 to equal 3", () => {
  expect(1 + 2).toBe(3);
});
