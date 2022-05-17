/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Validator } from "../stakeibc/validator";
import { ICAAccount } from "../stakeibc/ica_account";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** next id: 14 */
export interface HostZone {
  id: number;
  chainId: string;
  portId: string;
  channelId: string;
  connectionID: string;
  validators: Validator[];
  delegationAccounts: ICAAccount[];
  feeAccount: string;
  baseDenom: string;
  stDenom: string;
  totalDelegatorDelegations: string;
  totalAllBalances: string;
  totalOutstandingRewards: string;
}

const baseHostZone: object = {
  id: 0,
  chainId: "",
  portId: "",
  channelId: "",
  connectionID: "",
  feeAccount: "",
  baseDenom: "",
  stDenom: "",
  totalDelegatorDelegations: "",
  totalAllBalances: "",
  totalOutstandingRewards: "",
};

export const HostZone = {
  encode(message: HostZone, writer: Writer = Writer.create()): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    if (message.chainId !== "") {
      writer.uint32(18).string(message.chainId);
    }
    if (message.portId !== "") {
      writer.uint32(26).string(message.portId);
    }
    if (message.channelId !== "") {
      writer.uint32(34).string(message.channelId);
    }
    if (message.connectionID !== "") {
      writer.uint32(42).string(message.connectionID);
    }
    for (const v of message.validators) {
      Validator.encode(v!, writer.uint32(50).fork()).ldelim();
    }
    for (const v of message.delegationAccounts) {
      ICAAccount.encode(v!, writer.uint32(58).fork()).ldelim();
    }
    if (message.feeAccount !== "") {
      writer.uint32(66).string(message.feeAccount);
    }
    if (message.baseDenom !== "") {
      writer.uint32(74).string(message.baseDenom);
    }
    if (message.stDenom !== "") {
      writer.uint32(82).string(message.stDenom);
    }
    if (message.totalDelegatorDelegations !== "") {
      writer.uint32(90).string(message.totalDelegatorDelegations);
    }
    if (message.totalAllBalances !== "") {
      writer.uint32(98).string(message.totalAllBalances);
    }
    if (message.totalOutstandingRewards !== "") {
      writer.uint32(106).string(message.totalOutstandingRewards);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): HostZone {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.delegationAccounts = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.chainId = reader.string();
          break;
        case 3:
          message.portId = reader.string();
          break;
        case 4:
          message.channelId = reader.string();
          break;
        case 5:
          message.connectionID = reader.string();
          break;
        case 6:
          message.validators.push(Validator.decode(reader, reader.uint32()));
          break;
        case 7:
          message.delegationAccounts.push(
            ICAAccount.decode(reader, reader.uint32())
          );
          break;
        case 8:
          message.feeAccount = reader.string();
          break;
        case 9:
          message.baseDenom = reader.string();
          break;
        case 10:
          message.stDenom = reader.string();
          break;
        case 11:
          message.totalDelegatorDelegations = reader.string();
          break;
        case 12:
          message.totalAllBalances = reader.string();
          break;
        case 13:
          message.totalOutstandingRewards = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): HostZone {
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.delegationAccounts = [];
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.chainId !== undefined && object.chainId !== null) {
      message.chainId = String(object.chainId);
    } else {
      message.chainId = "";
    }
    if (object.portId !== undefined && object.portId !== null) {
      message.portId = String(object.portId);
    } else {
      message.portId = "";
    }
    if (object.channelId !== undefined && object.channelId !== null) {
      message.channelId = String(object.channelId);
    } else {
      message.channelId = "";
    }
    if (object.connectionID !== undefined && object.connectionID !== null) {
      message.connectionID = String(object.connectionID);
    } else {
      message.connectionID = "";
    }
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromJSON(e));
      }
    }
    if (
      object.delegationAccounts !== undefined &&
      object.delegationAccounts !== null
    ) {
      for (const e of object.delegationAccounts) {
        message.delegationAccounts.push(ICAAccount.fromJSON(e));
      }
    }
    if (object.feeAccount !== undefined && object.feeAccount !== null) {
      message.feeAccount = String(object.feeAccount);
    } else {
      message.feeAccount = "";
    }
    if (object.baseDenom !== undefined && object.baseDenom !== null) {
      message.baseDenom = String(object.baseDenom);
    } else {
      message.baseDenom = "";
    }
    if (object.stDenom !== undefined && object.stDenom !== null) {
      message.stDenom = String(object.stDenom);
    } else {
      message.stDenom = "";
    }
    if (
      object.totalDelegatorDelegations !== undefined &&
      object.totalDelegatorDelegations !== null
    ) {
      message.totalDelegatorDelegations = String(
        object.totalDelegatorDelegations
      );
    } else {
      message.totalDelegatorDelegations = "";
    }
    if (
      object.totalAllBalances !== undefined &&
      object.totalAllBalances !== null
    ) {
      message.totalAllBalances = String(object.totalAllBalances);
    } else {
      message.totalAllBalances = "";
    }
    if (
      object.totalOutstandingRewards !== undefined &&
      object.totalOutstandingRewards !== null
    ) {
      message.totalOutstandingRewards = String(object.totalOutstandingRewards);
    } else {
      message.totalOutstandingRewards = "";
    }
    return message;
  },

  toJSON(message: HostZone): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    message.chainId !== undefined && (obj.chainId = message.chainId);
    message.portId !== undefined && (obj.portId = message.portId);
    message.channelId !== undefined && (obj.channelId = message.channelId);
    message.connectionID !== undefined &&
      (obj.connectionID = message.connectionID);
    if (message.validators) {
      obj.validators = message.validators.map((e) =>
        e ? Validator.toJSON(e) : undefined
      );
    } else {
      obj.validators = [];
    }
    if (message.delegationAccounts) {
      obj.delegationAccounts = message.delegationAccounts.map((e) =>
        e ? ICAAccount.toJSON(e) : undefined
      );
    } else {
      obj.delegationAccounts = [];
    }
    message.feeAccount !== undefined && (obj.feeAccount = message.feeAccount);
    message.baseDenom !== undefined && (obj.baseDenom = message.baseDenom);
    message.stDenom !== undefined && (obj.stDenom = message.stDenom);
    message.totalDelegatorDelegations !== undefined &&
      (obj.totalDelegatorDelegations = message.totalDelegatorDelegations);
    message.totalAllBalances !== undefined &&
      (obj.totalAllBalances = message.totalAllBalances);
    message.totalOutstandingRewards !== undefined &&
      (obj.totalOutstandingRewards = message.totalOutstandingRewards);
    return obj;
  },

  fromPartial(object: DeepPartial<HostZone>): HostZone {
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.delegationAccounts = [];
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.chainId !== undefined && object.chainId !== null) {
      message.chainId = object.chainId;
    } else {
      message.chainId = "";
    }
    if (object.portId !== undefined && object.portId !== null) {
      message.portId = object.portId;
    } else {
      message.portId = "";
    }
    if (object.channelId !== undefined && object.channelId !== null) {
      message.channelId = object.channelId;
    } else {
      message.channelId = "";
    }
    if (object.connectionID !== undefined && object.connectionID !== null) {
      message.connectionID = object.connectionID;
    } else {
      message.connectionID = "";
    }
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromPartial(e));
      }
    }
    if (
      object.delegationAccounts !== undefined &&
      object.delegationAccounts !== null
    ) {
      for (const e of object.delegationAccounts) {
        message.delegationAccounts.push(ICAAccount.fromPartial(e));
      }
    }
    if (object.feeAccount !== undefined && object.feeAccount !== null) {
      message.feeAccount = object.feeAccount;
    } else {
      message.feeAccount = "";
    }
    if (object.baseDenom !== undefined && object.baseDenom !== null) {
      message.baseDenom = object.baseDenom;
    } else {
      message.baseDenom = "";
    }
    if (object.stDenom !== undefined && object.stDenom !== null) {
      message.stDenom = object.stDenom;
    } else {
      message.stDenom = "";
    }
    if (
      object.totalDelegatorDelegations !== undefined &&
      object.totalDelegatorDelegations !== null
    ) {
      message.totalDelegatorDelegations = object.totalDelegatorDelegations;
    } else {
      message.totalDelegatorDelegations = "";
    }
    if (
      object.totalAllBalances !== undefined &&
      object.totalAllBalances !== null
    ) {
      message.totalAllBalances = object.totalAllBalances;
    } else {
      message.totalAllBalances = "";
    }
    if (
      object.totalOutstandingRewards !== undefined &&
      object.totalOutstandingRewards !== null
    ) {
      message.totalOutstandingRewards = object.totalOutstandingRewards;
    } else {
      message.totalOutstandingRewards = "";
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
