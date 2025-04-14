package pluginmanager_service

import (
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// RateLimiterFromProto converts the proto format RateLimiterDefinition into a Defintion
func RateLimiterFromProto(p *proto.RateLimiterDefinition, pluginImageRef, pluginInstance string) (*plugin.RateLimiter, error) {
	var res = &plugin.RateLimiter{
		Name:  p.Name,
		Scope: p.Scope,
	}
	if p.FillRate != 0 {
		res.FillRate = &p.FillRate
		res.BucketSize = &p.BucketSize
	}
	if p.MaxConcurrency != 0 {
		res.MaxConcurrency = &p.MaxConcurrency
	}
	if p.Where != "" {
		res.Where = &p.Where
	}
	if res.Scope == nil {
		res.Scope = []string{}
	}
	// set ImageRef and Plugin fields
	res.SetPluginImageRef(pluginImageRef)
	res.PluginInstance = pluginInstance
	return res, nil
}

func RateLimiterAsProto(l *plugin.RateLimiter) *proto.RateLimiterDefinition {
	res := &proto.RateLimiterDefinition{
		Name:  l.Name,
		Scope: l.Scope,
	}
	if l.MaxConcurrency != nil {
		res.MaxConcurrency = *l.MaxConcurrency
	}
	if l.BucketSize != nil {
		res.BucketSize = *l.BucketSize
	}
	if l.FillRate != nil {
		res.FillRate = *l.FillRate
	}
	if l.Where != nil {
		res.Where = *l.Where
	}

	return res
}
