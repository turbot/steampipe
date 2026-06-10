package db_local

// Pins the on-disk JSON schema of the migration status files
// (public-migration-status.json / data-tank-migration-status.json) - the
// orchestrator-facing contract surface. The migration suites assert outcomes
// via the in-memory migrationResult; nothing else parses the file that
// external consumers read. This test exists because unexported
// partitionFailure fields once made every failed_partitions entry serialise
// as an empty object, invisibly to all outcome-level tests.
//
// If a field name here has to change, the orchestration contract document
// and the Pipes-side poller change with it - that is the point of the test.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationStatusJSONSchema(t *testing.T) {
	statusPath := filepath.Join(t.TempDir(), "status.json")

	res := migrationResult{
		tierReached:        dtRestoreTier3PerTable,
		reservedWordRouted: true,
		oldClusterRetained: true,
		committed:          false,
		dumpPath:           "/some/dump/dir",
		partitionFailures: []partitionFailure{{
			ParentSchema: "tank",
			ParentTable:  "aws_s3_bucket",
			PartSchema:   "tank-parts",
			PartTable:    "part_x-20260610000000",
			Reason:       "copy failed: disk full",
		}},
	}
	writeDataTankStatus(statusPath, res, "1 partition(s) could not be migrated; original preserved on disk")

	raw, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("status file not written: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("status file is not valid JSON: %v\n%s", err, raw)
	}

	// Top-level fields the orchestrator keys on, by exact name.
	for field, want := range map[string]any{
		"committed":            false,
		"tier_reached":         float64(dtRestoreTier3PerTable),
		"reserved_word_routed": true,
		"old_cluster_retained": true,
		"retained_dump_path":   "/some/dump/dir",
		"message":              "1 partition(s) could not be migrated; original preserved on disk",
	} {
		if got[field] != want {
			t.Errorf("field %q = %v, want %v", field, got[field], want)
		}
	}

	// The needs-help list must carry its content - empty objects here mean the
	// orchestrator cannot tell WHICH partitions need manual migration.
	parts, ok := got["failed_partitions"].([]any)
	if !ok || len(parts) != 1 {
		t.Fatalf("failed_partitions = %v, want exactly one entry", got["failed_partitions"])
	}
	entry, ok := parts[0].(map[string]any)
	if !ok {
		t.Fatalf("failed_partitions[0] is not an object: %v", parts[0])
	}
	for field, want := range map[string]string{
		"parent_schema": "tank",
		"parent_table":  "aws_s3_bucket",
		"part_schema":   "tank-parts",
		"part_table":    "part_x-20260610000000",
		"reason":        "copy failed: disk full",
	} {
		if entry[field] != want {
			t.Errorf("failed_partitions[0].%s = %v, want %q", field, entry[field], want)
		}
	}
}

// A success-shaped write must read back as committed with no failure noise.
func TestMigrationStatusJSONSchema_Success(t *testing.T) {
	statusPath := filepath.Join(t.TempDir(), "status.json")
	writeDataTankStatus(statusPath, migrationResult{
		tierReached: dtRestoreTier1Parallel,
		committed:   true,
		dumpPath:    "/some/dump/dir",
	}, "")

	raw, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("status file not written: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("status file is not valid JSON: %v\n%s", err, raw)
	}
	if got["committed"] != true {
		t.Errorf("committed = %v, want true", got["committed"])
	}
	if got["tier_reached"] != float64(dtRestoreTier1Parallel) {
		t.Errorf("tier_reached = %v, want %v", got["tier_reached"], float64(dtRestoreTier1Parallel))
	}
	if _, present := got["failed_partitions"]; present {
		t.Errorf("failed_partitions should be omitted on success, got %v", got["failed_partitions"])
	}
}
