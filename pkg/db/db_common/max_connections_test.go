package db_common

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func skipIfIntTooSmall(t *testing.T, value int64) {
	t.Helper()
	if strconv.IntSize < 64 && (value > math.MaxInt32 || value < math.MinInt32) {
		t.Skip("value does not fit in an int on this platform")
	}
}

// setMaxParallel sets the max-parallel setting for one test; a nil value leaves it unset
func setMaxParallel(t *testing.T, value *int64) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	if value != nil {
		viper.Set(pconstants.ArgMaxParallel, *value)
	}
}

func TestMaxDbConnections_AcceptsValuesThatFit(t *testing.T) {
	tests := []struct {
		name  string
		value *int64
		want  int32
	}{
		{name: "unset uses the default", want: int32(constants.DefaultMaxConnections)},
		{name: "10", value: ptr(10), want: 10},
		{name: "200", value: ptr(200), want: 200},
		{name: "largest allowed", value: ptr(math.MaxInt32), want: math.MaxInt32},
		{name: "smallest negative that fits", value: ptr(math.MinInt32), want: math.MinInt32},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMaxParallel(t, tt.value)
			got, err := MaxDbConnections()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMaxDbConnections_RejectsValuesThatDoNotFit(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{name: "one over the largest", value: math.MaxInt32 + 1},
		{name: "3000000000", value: 3000000000},
		{name: "negative that would wrap to a positive", value: -3000000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipIfIntTooSmall(t, tt.value)
			setMaxParallel(t, &tt.value)
			got, err := MaxDbConnections()
			if err == nil {
				t.Fatalf("expected an error, got %d", got)
			}
			// the error names the setting in each place it can be set, and the largest value
			for _, want := range []string{"max-parallel", "STEAMPIPE_MAX_PARALLEL", "max_parallel", "2147483647"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %q", err.Error(), want)
				}
			}
		})
	}
}

func ptr(v int64) *int64 { return &v }
