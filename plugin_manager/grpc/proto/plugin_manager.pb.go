// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: plugin_manager.proto

package proto

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type GetPluginRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Connection string `protobuf:"bytes,1,opt,name=connection,proto3" json:"connection,omitempty"`
}

func (x *GetPluginRequest) Reset() {
	*x = GetPluginRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_manager_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPluginRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPluginRequest) ProtoMessage() {}

func (x *GetPluginRequest) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_manager_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPluginRequest.ProtoReflect.Descriptor instead.
func (*GetPluginRequest) Descriptor() ([]byte, []int) {
	return file_plugin_manager_proto_rawDescGZIP(), []int{0}
}

func (x *GetPluginRequest) GetConnection() string {
	if x != nil {
		return x.Connection
	}
	return ""
}

type GetPluginResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Protocol        string   `protobuf:"bytes,1,opt,name=Protocol,proto3" json:"Protocol,omitempty"`
	ProtocolVersion int64    `protobuf:"varint,2,opt,name=ProtocolVersion,proto3" json:"ProtocolVersion,omitempty"`
	Addr            *NetAddr `protobuf:"bytes,3,opt,name=Addr,proto3" json:"Addr,omitempty"`
	Pid             int64    `protobuf:"varint,4,opt,name=Pid,proto3" json:"Pid,omitempty"`
}

func (x *GetPluginResponse) Reset() {
	*x = GetPluginResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_manager_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPluginResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPluginResponse) ProtoMessage() {}

func (x *GetPluginResponse) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_manager_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPluginResponse.ProtoReflect.Descriptor instead.
func (*GetPluginResponse) Descriptor() ([]byte, []int) {
	return file_plugin_manager_proto_rawDescGZIP(), []int{1}
}

func (x *GetPluginResponse) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *GetPluginResponse) GetProtocolVersion() int64 {
	if x != nil {
		return x.ProtocolVersion
	}
	return 0
}

func (x *GetPluginResponse) GetAddr() *NetAddr {
	if x != nil {
		return x.Addr
	}
	return nil
}

func (x *GetPluginResponse) GetPid() int64 {
	if x != nil {
		return x.Pid
	}
	return 0
}

type NetAddr struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Network string `protobuf:"bytes,1,opt,name=Network,proto3" json:"Network,omitempty"` // name of the network (for example, "tcp", "udp")
	Address string `protobuf:"bytes,2,opt,name=Address,proto3" json:"Address,omitempty"` // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
}

func (x *NetAddr) Reset() {
	*x = NetAddr{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_manager_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetAddr) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetAddr) ProtoMessage() {}

func (x *NetAddr) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_manager_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetAddr.ProtoReflect.Descriptor instead.
func (*NetAddr) Descriptor() ([]byte, []int) {
	return file_plugin_manager_proto_rawDescGZIP(), []int{2}
}

func (x *NetAddr) GetNetwork() string {
	if x != nil {
		return x.Network
	}
	return ""
}

func (x *NetAddr) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

var File_plugin_manager_proto protoreflect.FileDescriptor

var file_plugin_manager_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x32, 0x0a,
	0x10, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x22, 0x8f, 0x01, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x12, 0x28, 0x0a, 0x0f, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a,
	0x04, 0x41, 0x64, 0x64, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x52, 0x04, 0x41, 0x64, 0x64,
	0x72, 0x12, 0x10, 0x0a, 0x03, 0x50, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03,
	0x50, 0x69, 0x64, 0x22, 0x3d, 0x0a, 0x07, 0x4e, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x12, 0x18,
	0x0a, 0x07, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x18, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x32, 0x51, 0x0a, 0x0d, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4d, 0x61, 0x6e, 0x61,
	0x67, 0x65, 0x72, 0x12, 0x40, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x12, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_plugin_manager_proto_rawDescOnce sync.Once
	file_plugin_manager_proto_rawDescData = file_plugin_manager_proto_rawDesc
)

func file_plugin_manager_proto_rawDescGZIP() []byte {
	file_plugin_manager_proto_rawDescOnce.Do(func() {
		file_plugin_manager_proto_rawDescData = protoimpl.X.CompressGZIP(file_plugin_manager_proto_rawDescData)
	})
	return file_plugin_manager_proto_rawDescData
}

var file_plugin_manager_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_plugin_manager_proto_goTypes = []interface{}{
	(*GetPluginRequest)(nil),  // 0: proto.GetPluginRequest
	(*GetPluginResponse)(nil), // 1: proto.GetPluginResponse
	(*NetAddr)(nil),           // 2: proto.NetAddr
}
var file_plugin_manager_proto_depIdxs = []int32{
	2, // 0: proto.GetPluginResponse.Addr:type_name -> proto.NetAddr
	0, // 1: proto.PluginManager.GetPlugin:input_type -> proto.GetPluginRequest
	1, // 2: proto.PluginManager.GetPlugin:output_type -> proto.GetPluginResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_plugin_manager_proto_init() }
func file_plugin_manager_proto_init() {
	if File_plugin_manager_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_plugin_manager_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPluginRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_manager_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPluginResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_manager_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetAddr); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_plugin_manager_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_plugin_manager_proto_goTypes,
		DependencyIndexes: file_plugin_manager_proto_depIdxs,
		MessageInfos:      file_plugin_manager_proto_msgTypes,
	}.Build()
	File_plugin_manager_proto = out.File
	file_plugin_manager_proto_rawDesc = nil
	file_plugin_manager_proto_goTypes = nil
	file_plugin_manager_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// PluginManagerClient is the client API for PluginManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PluginManagerClient interface {
	GetPlugin(ctx context.Context, in *GetPluginRequest, opts ...grpc.CallOption) (*GetPluginResponse, error)
}

type pluginManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginManagerClient(cc grpc.ClientConnInterface) PluginManagerClient {
	return &pluginManagerClient{cc}
}

func (c *pluginManagerClient) GetPlugin(ctx context.Context, in *GetPluginRequest, opts ...grpc.CallOption) (*GetPluginResponse, error) {
	out := new(GetPluginResponse)
	err := c.cc.Invoke(ctx, "/proto.PluginManager/GetPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PluginManagerServer is the server API for PluginManager service.
type PluginManagerServer interface {
	GetPlugin(context.Context, *GetPluginRequest) (*GetPluginResponse, error)
}

// UnimplementedPluginManagerServer can be embedded to have forward compatible implementations.
type UnimplementedPluginManagerServer struct {
}

func (*UnimplementedPluginManagerServer) GetPlugin(context.Context, *GetPluginRequest) (*GetPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPlugin not implemented")
}

func RegisterPluginManagerServer(s *grpc.Server, srv PluginManagerServer) {
	s.RegisterService(&_PluginManager_serviceDesc, srv)
}

func _PluginManager_GetPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginManagerServer).GetPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PluginManager/GetPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginManagerServer).GetPlugin(ctx, req.(*GetPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PluginManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.PluginManager",
	HandlerType: (*PluginManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPlugin",
			Handler:    _PluginManager_GetPlugin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plugin_manager.proto",
}
