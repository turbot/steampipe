package pluginmanager_service

import (
	"github.com/hashicorp/go-plugin"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

type runningPlugin struct {
	imageRef       string
	pluginInstance string
	client         *plugin.Client
	reattach       *pb.ReattachConfig
	initialized    chan struct{}
	failed         chan struct{}
	error          error
	// set (under the plugin manager lock) when the user rate limiters changed while the plugin was starting,
	// so startPlugin sends them once the plugin has started
	limitersChangedWhileStarting bool
}

// isStarting returns true if this is still the placeholder added by addRunningPlugin,
// i.e. the plugin process has not finished starting so there is no client or reattach config yet
func (p *runningPlugin) isStarting() bool {
	return p.client == nil || p.reattach == nil
}
