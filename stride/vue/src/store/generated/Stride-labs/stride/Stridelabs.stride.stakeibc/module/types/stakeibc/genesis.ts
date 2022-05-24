/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Params } from "../stakeibc/params";
import { ICAAccount } from "../stakeibc/ica_account";
import { HostZone } from "../stakeibc/host_zone";
import { DepositRecord } from "../stakeibc/deposit_record";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** GenesisState defines the stakeibc module's genesis state. */
export interface GenesisState {
  params: Params | undefined;
  port_id: string;
  /** list of zones that are registered by the protocol */
  iCAAccount: ICAAccount | undefined;
  hostZoneList: HostZone[];
  hostZoneCount: number;
  depositRecordList: DepositRecord[];
  /** this line is used by starport scaffolding # genesis/proto/state */
  depositRecordCount: number;
}

const baseGenesisState: object = {
  port_id: "",
  hostZoneCount: 0,
  depositRecordCount: 0,
};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    if (message.port_id !== "") {
      writer.uint32(18).string(message.port_id);
    }
    if (message.iCAAccount !== undefined) {
      ICAAccount.encode(message.iCAAccount, writer.uint32(34).fork()).ldelim();
    }
    for (const v of message.hostZoneList) {
      HostZone.encode(v!, writer.uint32(42).fork()).ldelim();
    }
    if (message.hostZoneCount !== 0) {
      writer.uint32(48).uint64(message.hostZoneCount);
    }
    for (const v of message.depositRecordList) {
      DepositRecord.encode(v!, writer.uint32(58).fork()).ldelim();
    }
    if (message.depositRecordCount !== 0) {
      writer.uint32(64).uint64(message.depositRecordCount);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.hostZoneList = [];
    message.depositRecordList = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        case 2:
          message.port_id = reader.string();
          break;
        case 4:
          message.iCAAccount = ICAAccount.decode(reader, reader.uint32());
          break;
        case 5:
          message.hostZoneList.push(HostZone.decode(reader, reader.uint32()));
          break;
        case 6:
          message.hostZoneCount = longToNumber(reader.uint64() as Long);
          break;
        case 7:
          message.depositRecordList.push(
            DepositRecord.decode(reader, reader.uint32())
          );
          break;
        case 8:
          message.depositRecordCount = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.hostZoneList = [];
    message.depositRecordList = [];
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    if (object.port_id !== undefined && object.port_id !== null) {
      message.port_id = String(object.port_id);
    } else {
      message.port_id = "";
    }
    if (object.iCAAccount !== undefined && object.iCAAccount !== null) {
      message.iCAAccount = ICAAccount.fromJSON(object.iCAAccount);
    } else {
      message.iCAAccount = undefined;
    }
    if (object.hostZoneList !== undefined && object.hostZoneList !== null) {
      for (const e of object.hostZoneList) {
        message.hostZoneList.push(HostZone.fromJSON(e));
      }
    }
    if (object.hostZoneCount !== undefined && object.hostZoneCount !== null) {
      message.hostZoneCount = Number(object.hostZoneCount);
    } else {
      message.hostZoneCount = 0;
    }
    if (
      object.depositRecordList !== undefined &&
      object.depositRecordList !== null
    ) {
      for (const e of object.depositRecordList) {
        message.depositRecordList.push(DepositRecord.fromJSON(e));
      }
    }
    if (
      object.depositRecordCount !== undefined &&
      object.depositRecordCount !== null
    ) {
      message.depositRecordCount = Number(object.depositRecordCount);
    } else {
      message.depositRecordCount = 0;
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    message.port_id !== undefined && (obj.port_id = message.port_id);
    message.iCAAccount !== undefined &&
      (obj.iCAAccount = message.iCAAccount
        ? ICAAccount.toJSON(message.iCAAccount)
        : undefined);
    if (message.hostZoneList) {
      obj.hostZoneList = message.hostZoneList.map((e) =>
        e ? HostZone.toJSON(e) : undefined
      );
    } else {
      obj.hostZoneList = [];
    }
    message.hostZoneCount !== undefined &&
      (obj.hostZoneCount = message.hostZoneCount);
    if (message.depositRecordList) {
      obj.depositRecordList = message.depositRecordList.map((e) =>
        e ? DepositRecord.toJSON(e) : undefined
      );
    } else {
      obj.depositRecordList = [];
    }
    message.depositRecordCount !== undefined &&
      (obj.depositRecordCount = message.depositRecordCount);
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.hostZoneList = [];
    message.depositRecordList = [];
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    if (object.port_id !== undefined && object.port_id !== null) {
      message.port_id = object.port_id;
    } else {
      message.port_id = "";
    }
    if (object.iCAAccount !== undefined && object.iCAAccount !== null) {
      message.iCAAccount = ICAAccount.fromPartial(object.iCAAccount);
    } else {
      message.iCAAccount = undefined;
    }
    if (object.hostZoneList !== undefined && object.hostZoneList !== null) {
      for (const e of object.hostZoneList) {
        message.hostZoneList.push(HostZone.fromPartial(e));
      }
    }
    if (object.hostZoneCount !== undefined && object.hostZoneCount !== null) {
      message.hostZoneCount = object.hostZoneCount;
    } else {
      message.hostZoneCount = 0;
    }
    if (
      object.depositRecordList !== undefined &&
      object.depositRecordList !== null
    ) {
      for (const e of object.depositRecordList) {
        message.depositRecordList.push(DepositRecord.fromPartial(e));
      }
    }
    if (
      object.depositRecordCount !== undefined &&
      object.depositRecordCount !== null
    ) {
      message.depositRecordCount = object.depositRecordCount;
    } else {
      message.depositRecordCount = 0;
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
