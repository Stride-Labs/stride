import {
  SigningStargateClient,
  QueryClient,
  AuthExtension,
  BankExtension,
  StakingExtension,
  TxExtension,
  IbcExtension,
} from "@cosmjs/stargate";

export type Chain = "stride" | "cosmoshub" | "osmosis";

export type CosmosClient = {
  address: string;
  denom: string;
  client: SigningStargateClient;
  query: QueryClient & AuthExtension & BankExtension & StakingExtension & IbcExtension & TxExtension;
};

export type ChainConfig = {
  chainName: string;
  chainId: string;
  hostDenom: string;
  hostDenomOnStride: string;
  stDenom: string;
  strdDenomOnHost: string;
  bechPrefix: string;
  coinType: number;
  connectionId: string;
  transferChannelId: string;
  rpcEndpoint: string;
};

export type ChainConfigs = Record<string, ChainConfig>;

export type AccountInfo = {
  name: string;
  mnemonic: string;
  address?: string;
};
export type Mnemonics = {
  admin: AccountInfo;
  faucet: AccountInfo;
  validators: AccountInfo[];
  relayers: AccountInfo[];
  users: AccountInfo[];
};
