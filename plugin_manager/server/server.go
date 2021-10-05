/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a simple gRPC server that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
//
// It implements the route guide service whose definition can be found in routeguide/route_guide.proto.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/examples/data"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-plugin"
	pb "github.com/turbot/steampipe/plugin_manager/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type pluginManager struct {
	plugins map[string]*plugin.ReattachConfig
}

GetPlugin(context.Context, *GetPluginRequest) (*GetPluginResponse, error) {
return nil, status.Errorf(codes.Unimplemented, "method GetPlugin not implemented")
}
// GetFeature returns the feature at the given point.
func (s *pluginManager)GetPlugin(context.Context, *pb.GetPluginRequest) (*pb.GetPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPlugin not implemented")
}


func newServer() *pluginManager {
	s := &pluginManager{routeNotes: make(map[string][]*pb.RouteNote)}
	s.loadFeatures(*jsonDBFile)
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPluginManagerServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
