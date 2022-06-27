package controldisplay

import (
	"testing"
)

func BenchmarkToCsvCell(b *testing.B) {
	for i := 0; i < b.N; i++ {
		toCsvCell(i)
	}
}
