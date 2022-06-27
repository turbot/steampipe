package controlexecute

import (
	"fmt"
	"testing"
	"time"

	"github.com/logrusorgru/aurora"
)

func TestGetNextColor(t *testing.T) {
	var startingCol uint8
	var startingRow uint8
	for startingCol = 16; startingCol <= 51; startingCol++ {
		for startingRow = 0; startingRow <= 5; startingRow++ {
			fmt.Printf("\nROW %d COL %d\n", startingRow, startingCol)

			g, err := NewDimensionColorGenerator(startingRow, startingCol)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 10; i++ {
				color := g.getNextColor()
				fmt.Printf("%s\n", aurora.Index(color, fmt.Sprintf("XXXXXXXXXXXXXXXXXX, color: %d", color)))
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	fmt.Println()

}
