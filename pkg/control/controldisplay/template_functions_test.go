package controldisplay

import (
	"testing"
)

func BenchmarkToCsvCell(b *testing.B) {
	// the factory is called once per render execution
	toCsvCell := toCSVCellFnFactory("|")
	for i := 0; i < b.N; i++ {
		toCsvCell(i)
	}
}
