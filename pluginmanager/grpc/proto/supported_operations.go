package proto

import (
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
)

func SupportedOperationsFromSdk(s *sdkproto.GetSupportedOperationsResponse) *SupportedOperations {
	return &SupportedOperations{
		QueryCache:          s.QueryCache,
		CacheStream:         s.CacheStream,
		MultipleConnections: s.MultipleConnections,
	}
}
