package constants

import (
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe/version"
)

func GetSteampipeMetadata() *proto.SteampipeMetadata {
	return &proto.SteampipeMetadata{SteampipeVersion: version.String()}
}
