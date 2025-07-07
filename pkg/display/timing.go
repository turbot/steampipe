package display

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func DisplayTiming(result *queryresult.Result, rowCount int) {
	// show timing
	timingResult := getTiming(result, rowCount)
	if viper.GetString(pconstants.ArgTiming) != pconstants.ArgOff && timingResult != nil {
		str := buildTimingString(timingResult)
		if viper.GetBool(pconstants.ConfigKeyInteractive) {
			fmt.Println(str)
		} else {
			fmt.Fprintln(os.Stderr, str)
		}
	}
}

func getTiming(result *queryresult.Result, count int) *queryresult.TimingResult {
	timingConfig := viper.GetString(pconstants.ArgTiming)

	if timingConfig == pconstants.ArgOff || timingConfig == "false" {
		return nil
	}
	// now we have iterated the rows, get the timing
	timingResult := <-result.Timing.Stream
	// set rows returned
	timingResult.RowsReturned = int64(count)

	if timingConfig != pconstants.ArgVerbose {
		timingResult.Scans = nil
	}
	return timingResult
}

func buildTimingString(timingResult *queryresult.TimingResult) string {
	var sb strings.Builder
	// large numbers should be formatted with commas
	p := message.NewPrinter(language.English)

	sb.WriteString(fmt.Sprintf("\nTime: %s.", getDurationString(timingResult.DurationMs, p)))
	sb.WriteString(p.Sprintf(" Rows returned: %d.", timingResult.RowsReturned))
	totalRowsFetched := timingResult.UncachedRowsFetched + timingResult.CachedRowsFetched
	if totalRowsFetched == 0 {
		// maybe there was an error retrieving timing - just display the basics
		return sb.String()
	}

	sb.WriteString(" Rows fetched: ")
	if totalRowsFetched == 0 {
		sb.WriteString("0")
	} else {

		// calculate the number of cached rows fetched

		sb.WriteString(p.Sprintf("%d", totalRowsFetched))

		// were all cached
		if timingResult.UncachedRowsFetched == 0 {
			sb.WriteString(" (cached)")
		} else if timingResult.CachedRowsFetched > 0 {
			sb.WriteString(p.Sprintf(" (%d cached)", timingResult.CachedRowsFetched))
		}
	}

	sb.WriteString(p.Sprintf(". Hydrate calls: %d.", timingResult.HydrateCalls))
	if timingResult.ScanCount > 1 {
		sb.WriteString(p.Sprintf(" Scans: %d.", timingResult.ScanCount))
	}
	if timingResult.ConnectionCount > 1 {
		sb.WriteString(p.Sprintf(" Connections: %d.", timingResult.ConnectionCount))
	}

	if viper.GetString(pconstants.ArgTiming) == pconstants.ArgVerbose && len(timingResult.Scans) > 0 {
		if err := getVerboseTimingString(&sb, p, timingResult); err != nil {
			log.Printf("[WARN] Error getting verbose timing: %v", err)
		}
	}

	return sb.String()
}

func getDurationString(durationMs int64, p *message.Printer) string {
	if durationMs < 500 {
		return p.Sprintf("%dms", durationMs)
	} else {
		seconds := float64(durationMs) / 1000
		return p.Sprintf("%.1fs", seconds)
	}
}

func getVerboseTimingString(sb *strings.Builder, p *message.Printer, timingResult *queryresult.TimingResult) error {
	scans := timingResult.Scans

	// keep track of empty scans and do not include them separately in scan list
	emptyScanCount := 0
	scanCount := 0
	// is this all scans or just the slowest
	if len(scans) == int(timingResult.ScanCount) {
		sb.WriteString("\n\nScans:\n")
	} else {
		sb.WriteString(fmt.Sprintf("\n\nSlowest %d scans:\n", len(scans)))
	}

	for _, scan := range scans {
		if scan.RowsFetched == 0 {
			emptyScanCount++
			continue
		}
		scanCount++

		cacheString := ""
		if scan.CacheHit {
			cacheString = " (cached)"
		}
		qualsString := formatQuals(scan)
		limitString := ""
		if scan.Limit != nil {
			limitString = p.Sprintf(" Limit: %d.", *scan.Limit)
		}

		timeString := getDurationString(scan.DurationMs, p)
		rowsFetchedString := p.Sprintf("%d", scan.RowsFetched)

		sb.WriteString(p.Sprintf("  %d) %s.%s: Time: %s. Fetched: %s%s. Hydrates: %d.%s%s\n", scanCount, scan.Table, scan.Connection, timeString, rowsFetchedString, cacheString, scan.HydrateCalls, qualsString, limitString))
	}
	if emptyScanCount > 0 {

		sb.WriteString(fmt.Sprintf("  %d…%d) Zero rows fetched.\n", scanCount+1, scanCount+emptyScanCount))
	}
	return nil
}

func formatQuals(scan *queryresult.ScanMetadataRow) string {
	if len(scan.Quals) == 0 {
		return ""
	}

	var b strings.Builder
	for _, qual := range scan.Quals {
		operator := qual.Operator
		valueStr := formatQualValue(qual.Value)

		if operator == "=" {

			// Use reflection to check if qual.Value is an array or a slice
			val := reflect.ValueOf(qual.Value)

			if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
				// Change operator to IN if it was "=" and the value is an array or slice
				if operator == "=" {
					operator = " IN "
				}

				// Build the string of array elements
				valueElements := make([]string, val.Len())
				for i := 0; i < val.Len(); i++ {
					valueElements[i] = fmt.Sprintf("%s", formatQualValue(val.Index(i).Interface()))
				}
				valueStr = fmt.Sprintf("(%s)", strings.Join(valueElements, ", "))
			} else {
				// Use the original value if it's not an array or slice
				valueStr = fmt.Sprintf("%v", qual.Value)
			}
		}

		b.WriteString(fmt.Sprintf("%s%s%s, ", qual.Column, operator, valueStr))
	}

	// Remove the trailing comma and space
	trimmedResult := strings.TrimRight(b.String(), ", ")

	return fmt.Sprintf(" Quals: %s.", trimmedResult)
}

func formatQualValue(val any) string {
	if str, ok := val.(string); ok {
		return fmt.Sprintf("'%s'", str)
	}
	return fmt.Sprintf("%v", val)
}
