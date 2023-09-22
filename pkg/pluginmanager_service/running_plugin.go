package pluginmanager_service

import (
	"github.com/hashicorp/go-plugin"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
)

type runningPlugin struct {
	imageRef       string
	pluginInstance string
	client         *plugin.Client
	reattach       *pb.ReattachConfig
	initialized    chan struct{}
	failed         chan struct{}
	error          error
}
