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

func TestMaxDbConnections(t *testing.T) {
	tests := []struct {
		name    string
		set     bool
		value   int64
		want    int32
		wantErr bool
	}{
		{name: "unset uses the default", set: false, want: int32(constants.DefaultMaxConnections)},
		{name: "10", set: true, value: 10, want: 10},
		{name: "200", set: true, value: 200, want: 200},
		{name: "largest allowed", set: true, value: math.MaxInt32, want: math.MaxInt32},
		{name: "one over the largest", set: true, value: math.MaxInt32 + 1, wantErr: true},
		{name: "3000000000", set: true, value: 3000000000, wantErr: true},
		{name: "smallest negative that fits", set: true, value: math.MinInt32, want: math.MinInt32},
		{name: "negative that would wrap to a positive", set: true, value: -3000000000, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strconv.IntSize < 64 && (tt.value > math.MaxInt32 || tt.value < math.MinInt32) {
				t.Skip("value does not fit in an int on this platform")
			}
			viper.Reset()
			t.Cleanup(viper.Reset)
			if tt.set {
				viper.Set(pconstants.ArgMaxParallel, tt.value)
			}

			got, err := MaxDbConnections()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %d", got)
				}
				for _, want := range []string{"--max-parallel", "STEAMPIPE_MAX_PARALLEL", "max_parallel", "2147483647"} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q does not mention %q", err.Error(), want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
