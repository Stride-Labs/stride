/* eslint-disable */
import { Params } from "../stakeibc/params";
import { Validator } from "../stakeibc/validator";
import { Delegation } from "../stakeibc/delegation";
import { MinValidatorRequirements } from "../stakeibc/min_validator_requirements";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** GenesisState defines the stakeibc module's genesis state. */
export interface GenesisState {
  params: Params | undefined;
  port_id: string;
  validator: Validator | undefined;
  delegation: Delegation | undefined;
  /** this line is used by starport scaffolding # genesis/proto/state */
  minValidatorRequirements: MinValidatorRequirements | undefined;
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
    if (message.validator !== undefined) {
      Validator.encode(message.validator, writer.uint32(26).fork()).ldelim();
    }
    if (message.delegation !== undefined) {
      Delegation.encode(message.delegation, writer.uint32(34).fork()).ldelim();
    }
    if (message.minValidatorRequirements !== undefined) {
      MinValidatorRequirements.encode(
        message.minValidatorRequirements,
        writer.uint32(42).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
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
          message.validator = Validator.decode(reader, reader.uint32());
          break;
        case 4:
          message.delegation = Delegation.decode(reader, reader.uint32());
          break;
        case 5:
          message.minValidatorRequirements = MinValidatorRequirements.decode(
            reader,
            reader.uint32()
          );
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
    if (object.validator !== undefined && object.validator !== null) {
      message.validator = Validator.fromJSON(object.validator);
    } else {
      message.validator = undefined;
    }
    if (object.delegation !== undefined && object.delegation !== null) {
      message.delegation = Delegation.fromJSON(object.delegation);
    } else {
      message.delegation = undefined;
    }
    if (
      object.minValidatorRequirements !== undefined &&
      object.minValidatorRequirements !== null
    ) {
      message.minValidatorRequirements = MinValidatorRequirements.fromJSON(
        object.minValidatorRequirements
      );
    } else {
      message.minValidatorRequirements = undefined;
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    message.port_id !== undefined && (obj.port_id = message.port_id);
    message.validator !== undefined &&
      (obj.validator = message.validator
        ? Validator.toJSON(message.validator)
        : undefined);
    message.delegation !== undefined &&
      (obj.delegation = message.delegation
        ? Delegation.toJSON(message.delegation)
        : undefined);
    message.minValidatorRequirements !== undefined &&
      (obj.minValidatorRequirements = message.minValidatorRequirements
        ? MinValidatorRequirements.toJSON(message.minValidatorRequirements)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
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
    if (object.validator !== undefined && object.validator !== null) {
      message.validator = Validator.fromPartial(object.validator);
    } else {
      message.validator = undefined;
    }
    if (object.delegation !== undefined && object.delegation !== null) {
      message.delegation = Delegation.fromPartial(object.delegation);
    } else {
      message.delegation = undefined;
    }
    if (
      object.minValidatorRequirements !== undefined &&
      object.minValidatorRequirements !== null
    ) {
      message.minValidatorRequirements = MinValidatorRequirements.fromPartial(
        object.minValidatorRequirements
      );
    } else {
      message.minValidatorRequirements = undefined;
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
