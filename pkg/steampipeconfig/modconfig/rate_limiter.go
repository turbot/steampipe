package modconfig

import (
	"github.com/turbot/go-kit/hcl_helpers"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

const (
	LimiterSourceConfig     = "config"
	LimiterSourcePlugin     = "plugin"
	LimiterStatusActive     = "active"
	LimiterStatusOverridden = "overridden"
)

type RateLimiter struct {
	Name            string                          `hcl:"name,label" db:"name"`
	BucketSize      *int64                          `hcl:"bucket_size,optional" db:"bucket_size"`
	FillRate        *float32                        `hcl:"fill_rate,optional" db:"fill_rate"`
	MaxConcurrency  *int64                          `hcl:"max_concurrency,optional" db:"max_concurrency"`
	Scope           []string                        `hcl:"scope,optional" db:"scope"`
	Where           *string                         `hcl:"where,optional" db:"where"`
	Plugin          string                          `db:"plugin"`
	PluginInstance  string                          `db:"plugin_instance"`
	FileName        *string                         `db:"file_name" json:"-"`
	StartLineNumber *int                            `db:"start_line_number"  json:"-"`
	EndLineNumber   *int                            `db:"end_line_number"  json:"-"`
	Status          string                          `db:"status"`
	Source          string                          `db:"source_type"`
	ImageRef        *ociinstaller.SteampipeImageRef `db:"-" json:"-"`
}

// RateLimiterFromProto converts the proto format RateLimiterDefinition into a Defintion
func RateLimiterFromProto(p *proto.RateLimiterDefinition, pluginImageRef, pluginInstance string) (*RateLimiter, error) {
	var res = &RateLimiter{
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
	res.setPluginImageRef(pluginImageRef)
	res.PluginInstance = pluginInstance
	return res, nil
}

func (l *RateLimiter) AsProto() *proto.RateLimiterDefinition {
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

func (l *RateLimiter) OnDecoded(block *hcl.Block) {
	limiterRange := hcl_helpers.BlockRange(block)
	l.FileName = &limiterRange.Filename
	l.StartLineNumber = &limiterRange.Start.Line
	l.EndLineNumber = &limiterRange.End.Line
	if l.Scope == nil {
		l.Scope = []string{}
	}
	l.Status = LimiterStatusActive
	l.Source = LimiterSourceConfig
}

func (l *RateLimiter) scopeString() string {
	scope := l.Scope
	sort.Strings(scope)
	return strings.Join(scope, "'")
}

func (l *RateLimiter) Equals(other *RateLimiter) bool {
	return l.Name == other.Name &&
		pointersHaveSameValue(l.BucketSize, other.BucketSize) &&
		pointersHaveSameValue(l.FillRate, other.FillRate) &&
		pointersHaveSameValue(l.MaxConcurrency, other.MaxConcurrency) &&
		pointersHaveSameValue(l.Where, other.Where) &&
		l.scopeString() == other.scopeString() &&
		l.Plugin == other.Plugin &&
		l.PluginInstance == other.PluginInstance &&
		l.Source == other.Source
}

func (l *RateLimiter) SetPlugin(plugin *Plugin) {
	l.PluginInstance = plugin.Instance
	l.setPluginImageRef(plugin.Alias)
}

func (l *RateLimiter) setPluginImageRef(alias string) {
	l.ImageRef = ociinstaller.NewSteampipeImageRef(alias)
	l.Plugin = l.ImageRef.DisplayImageRef()

}

func pointersHaveSameValue[T comparable](l, r *T) bool {
	if l == nil {
		return r == nil
	}
	if r == nil {
		return false
	}
	return *l == *r
}
