package proto

import (
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

func SupportedOperationsFromSdk(s *sdkproto.GetSupportedOperationsResponse) *SupportedOperations {
	return &SupportedOperations{
		QueryCache:          s.QueryCache,
		MultipleConnections: s.MultipleConnections,
		MessageStream:       s.MessageStream,
	}
}
