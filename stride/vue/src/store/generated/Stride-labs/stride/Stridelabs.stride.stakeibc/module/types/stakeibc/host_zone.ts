/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Validator } from "../stakeibc/validator";
import { ICAAccount } from "../stakeibc/ica_account";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** next id: 8 */
export interface HostZone {
  id: number;
  portId: string;
  channelId: string;
  validators: Validator[];
  blacklistedValidators: Validator[];
  rewardsAccount: ICAAccount[];
  feeAccount: ICAAccount[];
  BaseDenom: string;
}

const baseHostZone: object = {
  id: 0,
  portId: "",
  channelId: "",
  BaseDenom: "",
};

export const HostZone = {
  encode(message: HostZone, writer: Writer = Writer.create()): Writer {
    if (message.id !== 0) {
      writer.uint32(56).uint64(message.id);
    }
    if (message.portId !== "") {
      writer.uint32(10).string(message.portId);
    }
    if (message.channelId !== "") {
      writer.uint32(18).string(message.channelId);
    }
    for (const v of message.validators) {
      Validator.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.blacklistedValidators) {
      Validator.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    for (const v of message.rewardsAccount) {
      ICAAccount.encode(v!, writer.uint32(42).fork()).ldelim();
    }
    for (const v of message.feeAccount) {
      ICAAccount.encode(v!, writer.uint32(50).fork()).ldelim();
    }
    if (message.BaseDenom !== "") {
      writer.uint32(74).string(message.BaseDenom);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): HostZone {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.blacklistedValidators = [];
    message.rewardsAccount = [];
    message.feeAccount = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 7:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 1:
          message.portId = reader.string();
          break;
        case 2:
          message.channelId = reader.string();
          break;
        case 3:
          message.validators.push(Validator.decode(reader, reader.uint32()));
          break;
        case 4:
          message.blacklistedValidators.push(
            Validator.decode(reader, reader.uint32())
          );
          break;
        case 5:
          message.rewardsAccount.push(
            ICAAccount.decode(reader, reader.uint32())
          );
          break;
        case 6:
          message.feeAccount.push(ICAAccount.decode(reader, reader.uint32()));
          break;
        case 9:
          message.BaseDenom = reader.string();
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
    message.blacklistedValidators = [];
    message.rewardsAccount = [];
    message.feeAccount = [];
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
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
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromJSON(e));
      }
    }
    if (
      object.blacklistedValidators !== undefined &&
      object.blacklistedValidators !== null
    ) {
      for (const e of object.blacklistedValidators) {
        message.blacklistedValidators.push(Validator.fromJSON(e));
      }
    }
    if (object.rewardsAccount !== undefined && object.rewardsAccount !== null) {
      for (const e of object.rewardsAccount) {
        message.rewardsAccount.push(ICAAccount.fromJSON(e));
      }
    }
    if (object.feeAccount !== undefined && object.feeAccount !== null) {
      for (const e of object.feeAccount) {
        message.feeAccount.push(ICAAccount.fromJSON(e));
      }
    }
    if (object.BaseDenom !== undefined && object.BaseDenom !== null) {
      message.BaseDenom = String(object.BaseDenom);
    } else {
      message.BaseDenom = "";
    }
    return message;
  },

  toJSON(message: HostZone): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    message.portId !== undefined && (obj.portId = message.portId);
    message.channelId !== undefined && (obj.channelId = message.channelId);
    if (message.validators) {
      obj.validators = message.validators.map((e) =>
        e ? Validator.toJSON(e) : undefined
      );
    } else {
      obj.validators = [];
    }
    if (message.blacklistedValidators) {
      obj.blacklistedValidators = message.blacklistedValidators.map((e) =>
        e ? Validator.toJSON(e) : undefined
      );
    } else {
      obj.blacklistedValidators = [];
    }
    if (message.rewardsAccount) {
      obj.rewardsAccount = message.rewardsAccount.map((e) =>
        e ? ICAAccount.toJSON(e) : undefined
      );
    } else {
      obj.rewardsAccount = [];
    }
    if (message.feeAccount) {
      obj.feeAccount = message.feeAccount.map((e) =>
        e ? ICAAccount.toJSON(e) : undefined
      );
    } else {
      obj.feeAccount = [];
    }
    message.BaseDenom !== undefined && (obj.BaseDenom = message.BaseDenom);
    return obj;
  },

  fromPartial(object: DeepPartial<HostZone>): HostZone {
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.blacklistedValidators = [];
    message.rewardsAccount = [];
    message.feeAccount = [];
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
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
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromPartial(e));
      }
    }
    if (
      object.blacklistedValidators !== undefined &&
      object.blacklistedValidators !== null
    ) {
      for (const e of object.blacklistedValidators) {
        message.blacklistedValidators.push(Validator.fromPartial(e));
      }
    }
    if (object.rewardsAccount !== undefined && object.rewardsAccount !== null) {
      for (const e of object.rewardsAccount) {
        message.rewardsAccount.push(ICAAccount.fromPartial(e));
      }
    }
    if (object.feeAccount !== undefined && object.feeAccount !== null) {
      for (const e of object.feeAccount) {
        message.feeAccount.push(ICAAccount.fromPartial(e));
      }
    }
    if (object.BaseDenom !== undefined && object.BaseDenom !== null) {
      message.BaseDenom = object.BaseDenom;
    } else {
      message.BaseDenom = "";
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
