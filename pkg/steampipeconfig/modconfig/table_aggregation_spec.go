package modconfig

import (
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"strings"
)

type TableAggregationSpecs []*TableAggregationSpec

func (s TableAggregationSpecs) ToProto() []*sdkproto.TableAggregationSpec {
	res := make([]*sdkproto.TableAggregationSpec, len(s))
	for i, t := range s {
		res[i] = &sdkproto.TableAggregationSpec{
			Match:       t.Match,
			Connections: t.Connections,
		}
	}
	return res
}

func (s TableAggregationSpecs) Equals(other TableAggregationSpecs) bool {
	if len(s) != len(other) {
		return false
	}
	for i, s := range s {
		if !s.Equals(other[i]) {
			return false
		}
	}
	return true
}

type TableAggregationSpec struct {
	Match       string   `hcl:"match,optional"`
	Connections []string `hcl:"connections"`
}

func (s TableAggregationSpec) Equals(other *TableAggregationSpec) bool {
	return s.Match == other.Match &&
		strings.Join(s.Connections, ",") == strings.Join(other.Connections, ",")
}
