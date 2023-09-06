package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"sort"
	"strings"
)

const (
	LimiterSourceConfig    = "config"
	LimiterSourcePlugin    = "plugin"
	LimiterStatusActive    = "active"
	LimiterStatusOverriden = "overridden"
)

type RateLimiter struct {
	Name            string   `hcl:"name,label" db:"name"`
	BucketSize      *int64   `hcl:"bucket_size,optional" db:"bucket_size"`
	FillRate        *float32 `hcl:"fill_rate,optional" db:"fill_rate"`
	MaxConcurrency  *int64   `hcl:"max_concurrency,optional" db:"max_concurrency"`
	Scope           []string `hcl:"scope,optional" db:"scope"`
	Where           *string  `hcl:"where,optional" db:"where"`
	Plugin          string   `db:"plugin"`
	FileName        *string  `db:"file_name"`
	StartLineNumber *int     `db:"start_line_number"`
	EndLineNumber   *int     `db:"end_line_number"`
	Status          string   `db:"status"`
	Source          string   `db:"source"`
}

// RateLimiterFromProto converts the proto format RateLimiterDefinition into a Defintion
func RateLimiterFromProto(p *proto.RateLimiterDefinition) (*RateLimiter, error) {
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
	l.FileName = &block.DefRange.Filename
	l.StartLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.Start.Line
	l.EndLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.End.Line
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
		l.BucketSize == other.BucketSize &&
		l.FillRate == other.FillRate &&
		l.scopeString() == other.scopeString() &&
		l.Where == other.Where
}
