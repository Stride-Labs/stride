// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgRegisterAccount } from "./types/stakeibc/tx";
import { MsgSubmitTx } from "./types/stakeibc/tx";
import { MsgLiquidStake } from "./types/stakeibc/tx";


const types = [
  ["/Stridelabs.stride.stakeibc.MsgRegisterAccount", MsgRegisterAccount],
  ["/Stridelabs.stride.stakeibc.MsgSubmitTx", MsgSubmitTx],
  ["/Stridelabs.stride.stakeibc.MsgLiquidStake", MsgLiquidStake],
  
];
export const MissingWalletError = new Error("wallet is required");

export const registry = new Registry(<any>types);

const defaultFee = {
  amount: [],
  gas: "200000",
};

interface TxClientOptions {
  addr: string
}

interface SignAndBroadcastOptions {
  fee: StdFee,
  memo?: string
}

const txClient = async (wallet: OfflineSigner, { addr: addr }: TxClientOptions = { addr: "http://localhost:26657" }) => {
  if (!wallet) throw MissingWalletError;
  let client;
  if (addr) {
    client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
  }else{
    client = await SigningStargateClient.offline( wallet, { registry });
  }
  const { address } = (await wallet.getAccounts())[0];

  return {
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }: SignAndBroadcastOptions = {fee: defaultFee, memo: ""}) => client.signAndBroadcast(address, msgs, fee,memo),
    msgRegisterAccount: (data: MsgRegisterAccount): EncodeObject => ({ typeUrl: "/Stridelabs.stride.stakeibc.MsgRegisterAccount", value: MsgRegisterAccount.fromPartial( data ) }),
    msgSubmitTx: (data: MsgSubmitTx): EncodeObject => ({ typeUrl: "/Stridelabs.stride.stakeibc.MsgSubmitTx", value: MsgSubmitTx.fromPartial( data ) }),
    msgLiquidStake: (data: MsgLiquidStake): EncodeObject => ({ typeUrl: "/Stridelabs.stride.stakeibc.MsgLiquidStake", value: MsgLiquidStake.fromPartial( data ) }),
    
  };
};

interface QueryClientOptions {
  addr: string
}

const queryClient = async ({ addr: addr }: QueryClientOptions = { addr: "http://localhost:1317" }) => {
  return new Api({ baseUrl: addr });
};

export {
  txClient,
  queryClient,
};
