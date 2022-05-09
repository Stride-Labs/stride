/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { Params } from "../stakeibc/params";
import { Validator } from "../stakeibc/validator";
import { Delegation } from "../stakeibc/delegation";
import { MinValidatorRequirements } from "../stakeibc/min_validator_requirements";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryGetValidatorRequest {}

export interface QueryGetValidatorResponse {
  Validator: Validator | undefined;
}

export interface QueryGetDelegationRequest {}

export interface QueryGetDelegationResponse {
  Delegation: Delegation | undefined;
}

export interface QueryGetMinValidatorRequirementsRequest {}

export interface QueryGetMinValidatorRequirementsResponse {
  MinValidatorRequirements: MinValidatorRequirements | undefined;
}

const baseQueryParamsRequest: object = {};

export const QueryParamsRequest = {
  encode(_: QueryParamsRequest, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },

  toJSON(_: QueryParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },
};

const baseQueryParamsResponse: object = {};

export const QueryParamsResponse = {
  encode(
    message: QueryParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

const baseQueryGetValidatorRequest: object = {};

export const QueryGetValidatorRequest = {
  encode(
    _: QueryGetValidatorRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetValidatorRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetValidatorRequest,
    } as QueryGetValidatorRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): QueryGetValidatorRequest {
    const message = {
      ...baseQueryGetValidatorRequest,
    } as QueryGetValidatorRequest;
    return message;
  },

  toJSON(_: QueryGetValidatorRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetValidatorRequest>
  ): QueryGetValidatorRequest {
    const message = {
      ...baseQueryGetValidatorRequest,
    } as QueryGetValidatorRequest;
    return message;
  },
};

const baseQueryGetValidatorResponse: object = {};

export const QueryGetValidatorResponse = {
  encode(
    message: QueryGetValidatorResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.Validator !== undefined) {
      Validator.encode(message.Validator, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetValidatorResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetValidatorResponse,
    } as QueryGetValidatorResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Validator = Validator.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetValidatorResponse {
    const message = {
      ...baseQueryGetValidatorResponse,
    } as QueryGetValidatorResponse;
    if (object.Validator !== undefined && object.Validator !== null) {
      message.Validator = Validator.fromJSON(object.Validator);
    } else {
      message.Validator = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetValidatorResponse): unknown {
    const obj: any = {};
    message.Validator !== undefined &&
      (obj.Validator = message.Validator
        ? Validator.toJSON(message.Validator)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetValidatorResponse>
  ): QueryGetValidatorResponse {
    const message = {
      ...baseQueryGetValidatorResponse,
    } as QueryGetValidatorResponse;
    if (object.Validator !== undefined && object.Validator !== null) {
      message.Validator = Validator.fromPartial(object.Validator);
    } else {
      message.Validator = undefined;
    }
    return message;
  },
};

const baseQueryGetDelegationRequest: object = {};

export const QueryGetDelegationRequest = {
  encode(
    _: QueryGetDelegationRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDelegationRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDelegationRequest,
    } as QueryGetDelegationRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): QueryGetDelegationRequest {
    const message = {
      ...baseQueryGetDelegationRequest,
    } as QueryGetDelegationRequest;
    return message;
  },

  toJSON(_: QueryGetDelegationRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetDelegationRequest>
  ): QueryGetDelegationRequest {
    const message = {
      ...baseQueryGetDelegationRequest,
    } as QueryGetDelegationRequest;
    return message;
  },
};

const baseQueryGetDelegationResponse: object = {};

export const QueryGetDelegationResponse = {
  encode(
    message: QueryGetDelegationResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.Delegation !== undefined) {
      Delegation.encode(message.Delegation, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDelegationResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDelegationResponse,
    } as QueryGetDelegationResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Delegation = Delegation.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDelegationResponse {
    const message = {
      ...baseQueryGetDelegationResponse,
    } as QueryGetDelegationResponse;
    if (object.Delegation !== undefined && object.Delegation !== null) {
      message.Delegation = Delegation.fromJSON(object.Delegation);
    } else {
      message.Delegation = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetDelegationResponse): unknown {
    const obj: any = {};
    message.Delegation !== undefined &&
      (obj.Delegation = message.Delegation
        ? Delegation.toJSON(message.Delegation)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetDelegationResponse>
  ): QueryGetDelegationResponse {
    const message = {
      ...baseQueryGetDelegationResponse,
    } as QueryGetDelegationResponse;
    if (object.Delegation !== undefined && object.Delegation !== null) {
      message.Delegation = Delegation.fromPartial(object.Delegation);
    } else {
      message.Delegation = undefined;
    }
    return message;
  },
};

const baseQueryGetMinValidatorRequirementsRequest: object = {};

export const QueryGetMinValidatorRequirementsRequest = {
  encode(
    _: QueryGetMinValidatorRequirementsRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetMinValidatorRequirementsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetMinValidatorRequirementsRequest,
    } as QueryGetMinValidatorRequirementsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): QueryGetMinValidatorRequirementsRequest {
    const message = {
      ...baseQueryGetMinValidatorRequirementsRequest,
    } as QueryGetMinValidatorRequirementsRequest;
    return message;
  },

  toJSON(_: QueryGetMinValidatorRequirementsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetMinValidatorRequirementsRequest>
  ): QueryGetMinValidatorRequirementsRequest {
    const message = {
      ...baseQueryGetMinValidatorRequirementsRequest,
    } as QueryGetMinValidatorRequirementsRequest;
    return message;
  },
};

const baseQueryGetMinValidatorRequirementsResponse: object = {};

export const QueryGetMinValidatorRequirementsResponse = {
  encode(
    message: QueryGetMinValidatorRequirementsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.MinValidatorRequirements !== undefined) {
      MinValidatorRequirements.encode(
        message.MinValidatorRequirements,
        writer.uint32(10).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetMinValidatorRequirementsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetMinValidatorRequirementsResponse,
    } as QueryGetMinValidatorRequirementsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.MinValidatorRequirements = MinValidatorRequirements.decode(
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

  fromJSON(object: any): QueryGetMinValidatorRequirementsResponse {
    const message = {
      ...baseQueryGetMinValidatorRequirementsResponse,
    } as QueryGetMinValidatorRequirementsResponse;
    if (
      object.MinValidatorRequirements !== undefined &&
      object.MinValidatorRequirements !== null
    ) {
      message.MinValidatorRequirements = MinValidatorRequirements.fromJSON(
        object.MinValidatorRequirements
      );
    } else {
      message.MinValidatorRequirements = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetMinValidatorRequirementsResponse): unknown {
    const obj: any = {};
    message.MinValidatorRequirements !== undefined &&
      (obj.MinValidatorRequirements = message.MinValidatorRequirements
        ? MinValidatorRequirements.toJSON(message.MinValidatorRequirements)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetMinValidatorRequirementsResponse>
  ): QueryGetMinValidatorRequirementsResponse {
    const message = {
      ...baseQueryGetMinValidatorRequirementsResponse,
    } as QueryGetMinValidatorRequirementsResponse;
    if (
      object.MinValidatorRequirements !== undefined &&
      object.MinValidatorRequirements !== null
    ) {
      message.MinValidatorRequirements = MinValidatorRequirements.fromPartial(
        object.MinValidatorRequirements
      );
    } else {
      message.MinValidatorRequirements = undefined;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a Validator by index. */
  Validator(
    request: QueryGetValidatorRequest
  ): Promise<QueryGetValidatorResponse>;
  /** Queries a Delegation by index. */
  Delegation(
    request: QueryGetDelegationRequest
  ): Promise<QueryGetDelegationResponse>;
  /** Queries a MinValidatorRequirements by index. */
  MinValidatorRequirements(
    request: QueryGetMinValidatorRequirementsRequest
  ): Promise<QueryGetMinValidatorRequirementsResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  Validator(
    request: QueryGetValidatorRequest
  ): Promise<QueryGetValidatorResponse> {
    const data = QueryGetValidatorRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "Validator",
      data
    );
    return promise.then((data) =>
      QueryGetValidatorResponse.decode(new Reader(data))
    );
  }

  Delegation(
    request: QueryGetDelegationRequest
  ): Promise<QueryGetDelegationResponse> {
    const data = QueryGetDelegationRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "Delegation",
      data
    );
    return promise.then((data) =>
      QueryGetDelegationResponse.decode(new Reader(data))
    );
  }

  MinValidatorRequirements(
    request: QueryGetMinValidatorRequirementsRequest
  ): Promise<QueryGetMinValidatorRequirementsResponse> {
    const data = QueryGetMinValidatorRequirementsRequest.encode(
      request
    ).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "MinValidatorRequirements",
      data
    );
    return promise.then((data) =>
      QueryGetMinValidatorRequirementsResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

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
