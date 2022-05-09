/* eslint-disable */
import { HostZone } from "../stakeibc/host_zone";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

export interface ICAAccount {
  address: string;
  balance: number;
  delegatedBalance: number;
  zone: HostZone | undefined;
}

const baseICAAccount: object = { address: "", balance: 0, delegatedBalance: 0 };

export const ICAAccount = {
  encode(message: ICAAccount, writer: Writer = Writer.create()): Writer {
    if (message.address !== "") {
      writer.uint32(10).string(message.address);
    }
    if (message.balance !== 0) {
      writer.uint32(16).int32(message.balance);
    }
    if (message.delegatedBalance !== 0) {
      writer.uint32(24).int32(message.delegatedBalance);
    }
    if (message.zone !== undefined) {
      HostZone.encode(message.zone, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): ICAAccount {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseICAAccount } as ICAAccount;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.address = reader.string();
          break;
        case 2:
          message.balance = reader.int32();
          break;
        case 3:
          message.delegatedBalance = reader.int32();
          break;
        case 4:
          message.zone = HostZone.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ICAAccount {
    const message = { ...baseICAAccount } as ICAAccount;
    if (object.address !== undefined && object.address !== null) {
      message.address = String(object.address);
    } else {
      message.address = "";
    }
    if (object.balance !== undefined && object.balance !== null) {
      message.balance = Number(object.balance);
    } else {
      message.balance = 0;
    }
    if (
      object.delegatedBalance !== undefined &&
      object.delegatedBalance !== null
    ) {
      message.delegatedBalance = Number(object.delegatedBalance);
    } else {
      message.delegatedBalance = 0;
    }
    if (object.zone !== undefined && object.zone !== null) {
      message.zone = HostZone.fromJSON(object.zone);
    } else {
      message.zone = undefined;
    }
    return message;
  },

  toJSON(message: ICAAccount): unknown {
    const obj: any = {};
    message.address !== undefined && (obj.address = message.address);
    message.balance !== undefined && (obj.balance = message.balance);
    message.delegatedBalance !== undefined &&
      (obj.delegatedBalance = message.delegatedBalance);
    message.zone !== undefined &&
      (obj.zone = message.zone ? HostZone.toJSON(message.zone) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<ICAAccount>): ICAAccount {
    const message = { ...baseICAAccount } as ICAAccount;
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = "";
    }
    if (object.balance !== undefined && object.balance !== null) {
      message.balance = object.balance;
    } else {
      message.balance = 0;
    }
    if (
      object.delegatedBalance !== undefined &&
      object.delegatedBalance !== null
    ) {
      message.delegatedBalance = object.delegatedBalance;
    } else {
      message.delegatedBalance = 0;
    }
    if (object.zone !== undefined && object.zone !== null) {
      message.zone = HostZone.fromPartial(object.zone);
    } else {
      message.zone = undefined;
    }
    return message;
  },
};

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
