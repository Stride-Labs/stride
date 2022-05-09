/* eslint-disable */
import { Validator } from "../stakeibc/validator";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

export interface Delegation {
  delegateAcctAddress: string;
  validator: Validator | undefined;
  amt: number;
}

const baseDelegation: object = { delegateAcctAddress: "", amt: 0 };

export const Delegation = {
  encode(message: Delegation, writer: Writer = Writer.create()): Writer {
    if (message.delegateAcctAddress !== "") {
      writer.uint32(10).string(message.delegateAcctAddress);
    }
    if (message.validator !== undefined) {
      Validator.encode(message.validator, writer.uint32(18).fork()).ldelim();
    }
    if (message.amt !== 0) {
      writer.uint32(24).int32(message.amt);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Delegation {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDelegation } as Delegation;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.delegateAcctAddress = reader.string();
          break;
        case 2:
          message.validator = Validator.decode(reader, reader.uint32());
          break;
        case 3:
          message.amt = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Delegation {
    const message = { ...baseDelegation } as Delegation;
    if (
      object.delegateAcctAddress !== undefined &&
      object.delegateAcctAddress !== null
    ) {
      message.delegateAcctAddress = String(object.delegateAcctAddress);
    } else {
      message.delegateAcctAddress = "";
    }
    if (object.validator !== undefined && object.validator !== null) {
      message.validator = Validator.fromJSON(object.validator);
    } else {
      message.validator = undefined;
    }
    if (object.amt !== undefined && object.amt !== null) {
      message.amt = Number(object.amt);
    } else {
      message.amt = 0;
    }
    return message;
  },

  toJSON(message: Delegation): unknown {
    const obj: any = {};
    message.delegateAcctAddress !== undefined &&
      (obj.delegateAcctAddress = message.delegateAcctAddress);
    message.validator !== undefined &&
      (obj.validator = message.validator
        ? Validator.toJSON(message.validator)
        : undefined);
    message.amt !== undefined && (obj.amt = message.amt);
    return obj;
  },

  fromPartial(object: DeepPartial<Delegation>): Delegation {
    const message = { ...baseDelegation } as Delegation;
    if (
      object.delegateAcctAddress !== undefined &&
      object.delegateAcctAddress !== null
    ) {
      message.delegateAcctAddress = object.delegateAcctAddress;
    } else {
      message.delegateAcctAddress = "";
    }
    if (object.validator !== undefined && object.validator !== null) {
      message.validator = Validator.fromPartial(object.validator);
    } else {
      message.validator = undefined;
    }
    if (object.amt !== undefined && object.amt !== null) {
      message.amt = object.amt;
    } else {
      message.amt = 0;
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
