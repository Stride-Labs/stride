// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stride/interchainquery/v1/messages.proto

package types

import (
	context "context"
	fmt "fmt"
	crypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// MsgSubmitQueryResponse represents a message type to fulfil a query request.
type MsgSubmitQueryResponse struct {
	ChainId     string           `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty" yaml:"chain_id"`
	QueryId     string           `protobuf:"bytes,2,opt,name=query_id,json=queryId,proto3" json:"query_id,omitempty" yaml:"query_id"`
	Result      []byte           `protobuf:"bytes,3,opt,name=result,proto3" json:"result,omitempty" yaml:"result"`
	ProofOps    *crypto.ProofOps `protobuf:"bytes,4,opt,name=proof_ops,json=proofOps,proto3" json:"proof_ops,omitempty" yaml:"proof_ops"`
	Height      int64            `protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty" yaml:"height"`
	FromAddress string           `protobuf:"bytes,6,opt,name=from_address,json=fromAddress,proto3" json:"from_address,omitempty"`
}

func (m *MsgSubmitQueryResponse) Reset()         { *m = MsgSubmitQueryResponse{} }
func (m *MsgSubmitQueryResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSubmitQueryResponse) ProtoMessage()    {}
func (*MsgSubmitQueryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_25adad4f8ed32400, []int{0}
}
func (m *MsgSubmitQueryResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitQueryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitQueryResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitQueryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitQueryResponse.Merge(m, src)
}
func (m *MsgSubmitQueryResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitQueryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitQueryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitQueryResponse proto.InternalMessageInfo

// MsgSubmitQueryResponseResponse defines the MsgSubmitQueryResponse response
// type.
type MsgSubmitQueryResponseResponse struct {
}

func (m *MsgSubmitQueryResponseResponse) Reset()         { *m = MsgSubmitQueryResponseResponse{} }
func (m *MsgSubmitQueryResponseResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSubmitQueryResponseResponse) ProtoMessage()    {}
func (*MsgSubmitQueryResponseResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_25adad4f8ed32400, []int{1}
}
func (m *MsgSubmitQueryResponseResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitQueryResponseResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitQueryResponseResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitQueryResponseResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitQueryResponseResponse.Merge(m, src)
}
func (m *MsgSubmitQueryResponseResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitQueryResponseResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitQueryResponseResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitQueryResponseResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgSubmitQueryResponse)(nil), "stride.interchainquery.v1.MsgSubmitQueryResponse")
	proto.RegisterType((*MsgSubmitQueryResponseResponse)(nil), "stride.interchainquery.v1.MsgSubmitQueryResponseResponse")
}

func init() {
	proto.RegisterFile("stride/interchainquery/v1/messages.proto", fileDescriptor_25adad4f8ed32400)
}

var fileDescriptor_25adad4f8ed32400 = []byte{
	// 513 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0x3f, 0x8f, 0xd3, 0x3e,
	0x18, 0xc7, 0xeb, 0xf6, 0xf7, 0xeb, 0xdd, 0xe5, 0x8a, 0x80, 0x5c, 0x85, 0x72, 0x05, 0x92, 0xca,
	0x0b, 0x05, 0x71, 0xb6, 0x5a, 0xc4, 0x70, 0x65, 0xa2, 0xdb, 0x49, 0x1c, 0x7f, 0xd2, 0x8d, 0xa5,
	0x4a, 0x1a, 0x9f, 0x6b, 0xa9, 0x89, 0x83, 0xed, 0x56, 0xd7, 0x95, 0x89, 0x11, 0x89, 0x85, 0xb1,
	0x2f, 0x02, 0x89, 0x91, 0x95, 0xf1, 0x04, 0x0b, 0x53, 0x85, 0x5a, 0x06, 0x58, 0xfb, 0x0a, 0x50,
	0xec, 0x04, 0xd0, 0x5d, 0x19, 0x98, 0xfc, 0xe4, 0xf9, 0x7e, 0x9e, 0x7f, 0xf1, 0x63, 0xab, 0x25,
	0x95, 0x60, 0x11, 0xc1, 0x2c, 0x51, 0x44, 0x0c, 0x47, 0x01, 0x4b, 0x5e, 0x4c, 0x88, 0x98, 0xe1,
	0x69, 0x1b, 0xc7, 0x44, 0xca, 0x80, 0x12, 0x89, 0x52, 0xc1, 0x15, 0xb7, 0xf7, 0x0d, 0x89, 0xce,
	0x91, 0x68, 0xda, 0x6e, 0xd4, 0x29, 0xa7, 0x5c, 0x53, 0x38, 0xb3, 0x4c, 0x40, 0x63, 0x7f, 0xc8,
	0x65, 0xcc, 0xe5, 0xc0, 0x08, 0xe6, 0x23, 0x97, 0x6e, 0x50, 0xce, 0xe9, 0x98, 0xe0, 0x20, 0x65,
	0x38, 0x48, 0x12, 0xae, 0x02, 0xc5, 0x78, 0x52, 0xa8, 0x37, 0x15, 0x49, 0x22, 0x22, 0x62, 0x96,
	0x28, 0x3c, 0x14, 0xb3, 0x54, 0x71, 0x9c, 0x0a, 0xce, 0x4f, 0x8c, 0x0c, 0x7f, 0x94, 0xad, 0x6b,
	0xc7, 0x92, 0xf6, 0x27, 0x61, 0xcc, 0xd4, 0xb3, 0xac, 0x07, 0x9f, 0xc8, 0x94, 0x27, 0x92, 0xd8,
	0xc8, 0xda, 0xd6, 0x9d, 0x0d, 0x58, 0xe4, 0x80, 0x26, 0x68, 0xed, 0xf4, 0xf6, 0xd6, 0x0b, 0xef,
	0xf2, 0x2c, 0x88, 0xc7, 0x5d, 0x58, 0x28, 0xd0, 0xdf, 0xd2, 0xe6, 0x51, 0x94, 0xf1, 0x7a, 0x88,
	0x8c, 0x2f, 0x9f, 0xe7, 0x0b, 0x05, 0xfa, 0x5b, 0xda, 0x3c, 0x8a, 0xec, 0xdb, 0x56, 0x55, 0x10,
	0x39, 0x19, 0x2b, 0xa7, 0xd2, 0x04, 0xad, 0x5a, 0xef, 0xea, 0x7a, 0xe1, 0x5d, 0x32, 0xb4, 0xf1,
	0x43, 0x3f, 0x07, 0xec, 0xc7, 0xd6, 0x8e, 0x6e, 0x7a, 0xc0, 0x53, 0xe9, 0xfc, 0xd7, 0x04, 0xad,
	0xdd, 0xce, 0x75, 0xf4, 0x7b, 0x30, 0x64, 0x06, 0x43, 0x4f, 0x33, 0xe6, 0x49, 0x2a, 0x7b, 0xf5,
	0xf5, 0xc2, 0xbb, 0x62, 0x52, 0xfd, 0x8a, 0x83, 0xfe, 0x76, 0x9a, 0xeb, 0x59, 0xe9, 0x11, 0x61,
	0x74, 0xa4, 0x9c, 0xff, 0x9b, 0xa0, 0x55, 0xf9, 0xb3, 0xb4, 0xf1, 0x43, 0x3f, 0x07, 0xec, 0x07,
	0x56, 0xed, 0x44, 0xf0, 0x78, 0x10, 0x44, 0x91, 0x20, 0x52, 0x3a, 0x55, 0x3d, 0x99, 0xf3, 0xe9,
	0xdd, 0x41, 0x3d, 0xbf, 0x85, 0x87, 0x46, 0xe9, 0x2b, 0xc1, 0x12, 0xea, 0xef, 0x66, 0x74, 0xee,
	0xea, 0xd6, 0x5e, 0xcd, 0xbd, 0xd2, 0xdb, 0xb9, 0x07, 0xbe, 0xcf, 0xbd, 0x12, 0x6c, 0x5a, 0xee,
	0xe6, 0x5f, 0x5d, 0x9c, 0x9d, 0x0f, 0xc0, 0xaa, 0x1c, 0x4b, 0x6a, 0xbf, 0x07, 0xd6, 0xde, 0xa6,
	0x2b, 0x69, 0xa3, 0xbf, 0xee, 0x0d, 0xda, 0x9c, 0xba, 0x71, 0xf8, 0xcf, 0x21, 0xc5, 0x09, 0x3b,
	0x2f, 0x3f, 0x7f, 0x7b, 0x53, 0xbe, 0xdb, 0x05, 0x77, 0xe0, 0xad, 0x0b, 0x3b, 0xad, 0x4e, 0xf1,
	0xb4, 0x1d, 0x12, 0x15, 0xb4, 0xb1, 0xd4, 0x39, 0xb4, 0xbb, 0xd7, 0xff, 0xb8, 0x74, 0xc1, 0xd9,
	0xd2, 0x05, 0x5f, 0x97, 0x2e, 0x78, 0xbd, 0x72, 0x4b, 0x67, 0x2b, 0xb7, 0xf4, 0x65, 0xe5, 0x96,
	0x9e, 0x1f, 0x52, 0xa6, 0x46, 0x93, 0x10, 0x0d, 0x79, 0x8c, 0xfb, 0xba, 0xa5, 0x83, 0x47, 0x41,
	0x28, 0x71, 0xfe, 0x66, 0xa6, 0x9d, 0xfb, 0xf8, 0xf4, 0x62, 0x95, 0x59, 0x4a, 0x64, 0x58, 0xd5,
	0xbb, 0x7a, 0xef, 0x67, 0x00, 0x00, 0x00, 0xff, 0xff, 0x08, 0x00, 0xa2, 0x15, 0x60, 0x03, 0x00,
	0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// SubmitQueryResponse defines a method for submit query responses.
	SubmitQueryResponse(ctx context.Context, in *MsgSubmitQueryResponse, opts ...grpc.CallOption) (*MsgSubmitQueryResponseResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SubmitQueryResponse(ctx context.Context, in *MsgSubmitQueryResponse, opts ...grpc.CallOption) (*MsgSubmitQueryResponseResponse, error) {
	out := new(MsgSubmitQueryResponseResponse)
	err := c.cc.Invoke(ctx, "/stride.interchainquery.v1.Msg/SubmitQueryResponse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// SubmitQueryResponse defines a method for submit query responses.
	SubmitQueryResponse(context.Context, *MsgSubmitQueryResponse) (*MsgSubmitQueryResponseResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) SubmitQueryResponse(ctx context.Context, req *MsgSubmitQueryResponse) (*MsgSubmitQueryResponseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitQueryResponse not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_SubmitQueryResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitQueryResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitQueryResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stride.interchainquery.v1.Msg/SubmitQueryResponse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitQueryResponse(ctx, req.(*MsgSubmitQueryResponse))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stride.interchainquery.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitQueryResponse",
			Handler:    _Msg_SubmitQueryResponse_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "stride/interchainquery/v1/messages.proto",
}

func (m *MsgSubmitQueryResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitQueryResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitQueryResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.FromAddress) > 0 {
		i -= len(m.FromAddress)
		copy(dAtA[i:], m.FromAddress)
		i = encodeVarintMessages(dAtA, i, uint64(len(m.FromAddress)))
		i--
		dAtA[i] = 0x32
	}
	if m.Height != 0 {
		i = encodeVarintMessages(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x28
	}
	if m.ProofOps != nil {
		{
			size, err := m.ProofOps.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintMessages(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	if len(m.Result) > 0 {
		i -= len(m.Result)
		copy(dAtA[i:], m.Result)
		i = encodeVarintMessages(dAtA, i, uint64(len(m.Result)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.QueryId) > 0 {
		i -= len(m.QueryId)
		copy(dAtA[i:], m.QueryId)
		i = encodeVarintMessages(dAtA, i, uint64(len(m.QueryId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ChainId) > 0 {
		i -= len(m.ChainId)
		copy(dAtA[i:], m.ChainId)
		i = encodeVarintMessages(dAtA, i, uint64(len(m.ChainId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSubmitQueryResponseResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitQueryResponseResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitQueryResponseResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintMessages(dAtA []byte, offset int, v uint64) int {
	offset -= sovMessages(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgSubmitQueryResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ChainId)
	if l > 0 {
		n += 1 + l + sovMessages(uint64(l))
	}
	l = len(m.QueryId)
	if l > 0 {
		n += 1 + l + sovMessages(uint64(l))
	}
	l = len(m.Result)
	if l > 0 {
		n += 1 + l + sovMessages(uint64(l))
	}
	if m.ProofOps != nil {
		l = m.ProofOps.Size()
		n += 1 + l + sovMessages(uint64(l))
	}
	if m.Height != 0 {
		n += 1 + sovMessages(uint64(m.Height))
	}
	l = len(m.FromAddress)
	if l > 0 {
		n += 1 + l + sovMessages(uint64(l))
	}
	return n
}

func (m *MsgSubmitQueryResponseResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovMessages(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMessages(x uint64) (n int) {
	return sovMessages(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgSubmitQueryResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessages
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgSubmitQueryResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitQueryResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMessages
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMessages
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field QueryId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMessages
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMessages
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.QueryId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Result", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessages
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMessages
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Result = append(m.Result[:0], dAtA[iNdEx:postIndex]...)
			if m.Result == nil {
				m.Result = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProofOps", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMessages
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMessages
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ProofOps == nil {
				m.ProofOps = &crypto.ProofOps{}
			}
			if err := m.ProofOps.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FromAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMessages
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMessages
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FromAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessages(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMessages
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgSubmitQueryResponseResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessages
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgSubmitQueryResponseResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitQueryResponseResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipMessages(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMessages
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMessages(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMessages
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthMessages
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMessages
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMessages
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMessages        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMessages          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMessages = fmt.Errorf("proto: unexpected end of group")
)
