package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

var profiling = false

type timeLog struct {
	Time       time.Time
	Interval   time.Duration
	Cumulative time.Duration
	Operation  string
}

var Timing []timeLog

func shouldProfile() bool {
	profilingEnv, exists := os.LookupEnv("STEAMPIPE_PROFILE")
	if exists {
		return strings.ToUpper(profilingEnv) == "TRUE"
	}
	return profiling
}
func LogTime(operation string) {
	if !shouldProfile() {
		return
	}
	lastTimelogIdx := len(Timing) - 1
	var elapsed time.Duration
	var cumulative time.Duration
	if lastTimelogIdx >= 0 {
		elapsed = time.Since(Timing[lastTimelogIdx].Time)
		cumulative = time.Since(Timing[0].Time)

	}
	Timing = append(Timing, timeLog{time.Now(), elapsed, cumulative, operation})
}

func DisplayProfileData() {
	if shouldProfile() {
		fmt.Println("Timing")

		var data [][]string
		for _, logEntry := range Timing {
			var itemData []string
			itemData = append(itemData, logEntry.Operation)
			itemData = append(itemData, fmt.Sprintf("%s", logEntry.Interval))
			itemData = append(itemData, fmt.Sprintf("%s", logEntry.Cumulative))
			data = append(data, itemData)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Operation", "Elapsed", "Cumulative"})
		table.SetBorder(true)
		table.AppendBulk(data)
		table.SetAutoWrapText(false)
		table.Render()
	}

}
