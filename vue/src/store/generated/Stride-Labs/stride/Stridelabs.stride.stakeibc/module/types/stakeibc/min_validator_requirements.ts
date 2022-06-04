/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

export interface MinValidatorRequirements {
  commissionRate: number;
  uptime: number;
}

const baseMinValidatorRequirements: object = { commissionRate: 0, uptime: 0 };

export const MinValidatorRequirements = {
  encode(
    message: MinValidatorRequirements,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.commissionRate !== 0) {
      writer.uint32(8).int32(message.commissionRate);
    }
    if (message.uptime !== 0) {
      writer.uint32(16).int32(message.uptime);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MinValidatorRequirements {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMinValidatorRequirements,
    } as MinValidatorRequirements;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.commissionRate = reader.int32();
          break;
        case 2:
          message.uptime = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MinValidatorRequirements {
    const message = {
      ...baseMinValidatorRequirements,
    } as MinValidatorRequirements;
    if (object.commissionRate !== undefined && object.commissionRate !== null) {
      message.commissionRate = Number(object.commissionRate);
    } else {
      message.commissionRate = 0;
    }
    if (object.uptime !== undefined && object.uptime !== null) {
      message.uptime = Number(object.uptime);
    } else {
      message.uptime = 0;
    }
    return message;
  },

  toJSON(message: MinValidatorRequirements): unknown {
    const obj: any = {};
    message.commissionRate !== undefined &&
      (obj.commissionRate = message.commissionRate);
    message.uptime !== undefined && (obj.uptime = message.uptime);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MinValidatorRequirements>
  ): MinValidatorRequirements {
    const message = {
      ...baseMinValidatorRequirements,
    } as MinValidatorRequirements;
    if (object.commissionRate !== undefined && object.commissionRate !== null) {
      message.commissionRate = object.commissionRate;
    } else {
      message.commissionRate = 0;
    }
    if (object.uptime !== undefined && object.uptime !== null) {
      message.uptime = object.uptime;
    } else {
      message.uptime = 0;
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
