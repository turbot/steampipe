package modconfig

import (
	"fmt"
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
	LimiterStatusOverriden = "overriden"
)

type ResolvedRateLimiter struct {
	*RateLimiter
	Status string
	Source string
}

type RateLimiter struct {
	Name            string   `hcl:"name,label"`
	Plugin          string   `hcl:"plugin"`
	BucketSize      *int64   `hcl:"bucket_size,optional"`
	FillRate        *float32 `hcl:"fill_rate,optional"`
	MaxConcurrency  *int64   `hcl:"max_concurrency,optional"`
	Scope           []string `hcl:"scope"`
	Where           *string  `hcl:"where,optional"`
	QualifiedName   string
	FileName        *string
	StartLineNumber *int
	EndLineNumber   *int
}

func (l RateLimiter) scopeString() string {
	scope := l.Scope
	sort.Strings(scope)
	return strings.Join(scope, "'")
}

func (l RateLimiter) AsProto() *proto.RateLimiterDefinition {
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

func (l RateLimiter) Equals(other *RateLimiter) bool {
	return l.Name == other.Name &&
		l.BucketSize == other.BucketSize &&
		l.FillRate == other.FillRate &&
		l.scopeString() == other.scopeString() &&
		l.Where == other.Where
}

func (l RateLimiter) OnDecoded(block *hcl.Block) {
	l.FileName = &block.DefRange.Filename
	l.StartLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.Start.Line
	l.EndLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.End.Line
	l.QualifiedName = fmt.Sprintf("%s.%s", l.Plugin, l.Name)

}
