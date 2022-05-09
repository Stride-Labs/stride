/* eslint-disable */
import { Params } from "../stakeibc/params";
import { HostZone } from "../stakeibc/host_zone";
import { ICAAccount } from "../stakeibc/ica_account";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** GenesisState defines the stakeibc module's genesis state. */
export interface GenesisState {
  params: Params | undefined;
  port_id: string;
  /** list of zones that are registered by the protocol */
  hostZone: HostZone[];
  /** this line is used by starport scaffolding # genesis/proto/state */
  iCAAccount: ICAAccount | undefined;
}

const baseGenesisState: object = { port_id: "" };

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    if (message.port_id !== "") {
      writer.uint32(18).string(message.port_id);
    }
    for (const v of message.hostZone) {
      HostZone.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    if (message.iCAAccount !== undefined) {
      ICAAccount.encode(message.iCAAccount, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.hostZone = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        case 2:
          message.port_id = reader.string();
          break;
        case 3:
          message.hostZone.push(HostZone.decode(reader, reader.uint32()));
          break;
        case 4:
          message.iCAAccount = ICAAccount.decode(reader, reader.uint32());
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
    message.hostZone = [];
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
    if (object.hostZone !== undefined && object.hostZone !== null) {
      for (const e of object.hostZone) {
        message.hostZone.push(HostZone.fromJSON(e));
      }
    }
    if (object.iCAAccount !== undefined && object.iCAAccount !== null) {
      message.iCAAccount = ICAAccount.fromJSON(object.iCAAccount);
    } else {
      message.iCAAccount = undefined;
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    message.port_id !== undefined && (obj.port_id = message.port_id);
    if (message.hostZone) {
      obj.hostZone = message.hostZone.map((e) =>
        e ? HostZone.toJSON(e) : undefined
      );
    } else {
      obj.hostZone = [];
    }
    message.iCAAccount !== undefined &&
      (obj.iCAAccount = message.iCAAccount
        ? ICAAccount.toJSON(message.iCAAccount)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.hostZone = [];
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
    if (object.hostZone !== undefined && object.hostZone !== null) {
      for (const e of object.hostZone) {
        message.hostZone.push(HostZone.fromPartial(e));
      }
    }
    if (object.iCAAccount !== undefined && object.iCAAccount !== null) {
      message.iCAAccount = ICAAccount.fromPartial(object.iCAAccount);
    } else {
      message.iCAAccount = undefined;
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
