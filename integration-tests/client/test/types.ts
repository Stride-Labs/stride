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
  chainId: string;
  hostDenom: string;
  bechPrefix: string;
  coinType: number;
  connectionId: string;
  transferChannelId: string;
  rpcEndpoint: string;
};

export type ChainConfigs = Record<string, ChainConfig>;
