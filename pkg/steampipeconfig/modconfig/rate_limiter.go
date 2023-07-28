package modconfig

import (
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

type RateLimiter struct {
	Name            string   `hcl:"name,label"`
	Plugin          string   `hcl:"plugin"`
	BucketSize      int64    `hcl:"bucket_size"`
	FillRate        float32  `hcl:"fill_rate"`
	Scope           []string `hcl:"scope"`
	Where           string   `hcl:"where"`
	Status          string
	Source          string
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
	return &proto.RateLimiterDefinition{
		Name:       l.Name,
		FillRate:   l.FillRate,
		BucketSize: l.BucketSize,
		Scopes:     l.Scope,
		Where:      l.Where,
	}
}

func (l RateLimiter) Equals(other *RateLimiter) bool {
	return l.Name == other.Name &&
		l.BucketSize == other.BucketSize &&
		l.FillRate == other.FillRate &&
		l.scopeString() == other.scopeString() &&
		l.Where == other.Where
}
