// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: metrics/load.proto

package metrics // import "github.com/linkai-io/am/protocservices/metrics"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type LoadRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoadRequest) Reset()         { *m = LoadRequest{} }
func (m *LoadRequest) String() string { return proto.CompactTextString(m) }
func (*LoadRequest) ProtoMessage()    {}
func (*LoadRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_load_7181cb0cf0a2dbed, []int{0}
}
func (m *LoadRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LoadRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LoadRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *LoadRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoadRequest.Merge(dst, src)
}
func (m *LoadRequest) XXX_Size() int {
	return m.Size()
}
func (m *LoadRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_LoadRequest.DiscardUnknown(m)
}

var xxx_messageInfo_LoadRequest proto.InternalMessageInfo

type LoadResponse struct {
	Score                int64    `protobuf:"varint,1,opt,name=score,proto3" json:"score,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoadResponse) Reset()         { *m = LoadResponse{} }
func (m *LoadResponse) String() string { return proto.CompactTextString(m) }
func (*LoadResponse) ProtoMessage()    {}
func (*LoadResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_load_7181cb0cf0a2dbed, []int{1}
}
func (m *LoadResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LoadResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LoadResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *LoadResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoadResponse.Merge(dst, src)
}
func (m *LoadResponse) XXX_Size() int {
	return m.Size()
}
func (m *LoadResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_LoadResponse.DiscardUnknown(m)
}

var xxx_messageInfo_LoadResponse proto.InternalMessageInfo

func (m *LoadResponse) GetScore() int64 {
	if m != nil {
		return m.Score
	}
	return 0
}

func init() {
	proto.RegisterType((*LoadRequest)(nil), "metrics.LoadRequest")
	proto.RegisterType((*LoadResponse)(nil), "metrics.LoadResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// LoadReportClient is the client API for LoadReport service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type LoadReportClient interface {
	Load(ctx context.Context, in *LoadRequest, opts ...grpc.CallOption) (*LoadResponse, error)
}

type loadReportClient struct {
	cc *grpc.ClientConn
}

func NewLoadReportClient(cc *grpc.ClientConn) LoadReportClient {
	return &loadReportClient{cc}
}

func (c *loadReportClient) Load(ctx context.Context, in *LoadRequest, opts ...grpc.CallOption) (*LoadResponse, error) {
	out := new(LoadResponse)
	err := c.cc.Invoke(ctx, "/metrics.LoadReport/Load", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LoadReportServer is the server API for LoadReport service.
type LoadReportServer interface {
	Load(context.Context, *LoadRequest) (*LoadResponse, error)
}

func RegisterLoadReportServer(s *grpc.Server, srv LoadReportServer) {
	s.RegisterService(&_LoadReport_serviceDesc, srv)
}

func _LoadReport_Load_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LoadReportServer).Load(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metrics.LoadReport/Load",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LoadReportServer).Load(ctx, req.(*LoadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LoadReport_serviceDesc = grpc.ServiceDesc{
	ServiceName: "metrics.LoadReport",
	HandlerType: (*LoadReportServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Load",
			Handler:    _LoadReport_Load_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "metrics/load.proto",
}

func (m *LoadRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LoadRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *LoadResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LoadResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Score != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintLoad(dAtA, i, uint64(m.Score))
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintLoad(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *LoadRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *LoadResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Score != 0 {
		n += 1 + sovLoad(uint64(m.Score))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovLoad(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozLoad(x uint64) (n int) {
	return sovLoad(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *LoadRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLoad
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LoadRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LoadRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipLoad(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLoad
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *LoadResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLoad
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LoadResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LoadResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Score", wireType)
			}
			m.Score = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLoad
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Score |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLoad(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLoad
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLoad(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLoad
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
					return 0, ErrIntOverflowLoad
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLoad
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthLoad
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowLoad
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipLoad(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthLoad = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLoad   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("metrics/load.proto", fileDescriptor_load_7181cb0cf0a2dbed) }

var fileDescriptor_load_7181cb0cf0a2dbed = []byte{
	// 187 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xca, 0x4d, 0x2d, 0x29,
	0xca, 0x4c, 0x2e, 0xd6, 0xcf, 0xc9, 0x4f, 0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62,
	0x87, 0x8a, 0x29, 0xf1, 0x72, 0x71, 0xfb, 0xe4, 0x27, 0xa6, 0x04, 0xa5, 0x16, 0x96, 0xa6, 0x16,
	0x97, 0x28, 0xa9, 0x70, 0xf1, 0x40, 0xb8, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x42, 0x22, 0x5c,
	0xac, 0xc5, 0xc9, 0xf9, 0x45, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0xcc, 0x41, 0x10, 0x8e, 0x91,
	0x23, 0x17, 0x17, 0x44, 0x55, 0x41, 0x7e, 0x51, 0x89, 0x90, 0x31, 0x17, 0x0b, 0x88, 0x27, 0x24,
	0xa2, 0x07, 0x35, 0x54, 0x0f, 0xc9, 0x44, 0x29, 0x51, 0x34, 0x51, 0x88, 0xc1, 0x4e, 0x0e, 0x27,
	0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3, 0x8c, 0xc7, 0x72, 0x0c,
	0x51, 0x7a, 0xe9, 0x99, 0x25, 0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xfa, 0x39, 0x99, 0x79,
	0xd9, 0x89, 0x99, 0xba, 0x99, 0xf9, 0xfa, 0x89, 0xb9, 0xfa, 0x60, 0x17, 0x27, 0x17, 0xa7, 0x16,
	0x95, 0x65, 0x26, 0xa7, 0x16, 0xeb, 0x43, 0x8d, 0x4b, 0x62, 0x03, 0x8b, 0x1b, 0x03, 0x02, 0x00,
	0x00, 0xff, 0xff, 0xcf, 0x0b, 0x26, 0xc2, 0xdf, 0x00, 0x00, 0x00,
}