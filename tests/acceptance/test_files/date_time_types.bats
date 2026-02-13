load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# Test DATE, TIMESTAMP, TIMESTAMPTZ display formatting
# Verifies fix for issue #4450

@test "DATE displays without time component" {
  run steampipe query "select '1984-01-01'::date as date_val" --output json
  echo "$output" | jq -e '.rows[0].date_val == "1984-01-01"'
  assert_success
}

@test "DATE with table output" {
  run steampipe query "select '2024-02-29'::date as leap_date"
  assert_output --partial "2024-02-29"
  refute_output --partial "00:00:00"
}

@test "DATE NULL value" {
  run steampipe query "select null::date as null_date" --output json
  echo "$output" | jq -e '.rows[0].null_date == null'
  assert_success
}

@test "TIMESTAMPTZ displays with UTC timezone" {
  run steampipe query "select '1984-01-01T00:00:00Z'::timestamptz as tstz" --output json
  echo "$output" | jq -e '.rows[0].tstz == "1984-01-01T00:00:00Z"'
  assert_success
}

@test "TIMESTAMPTZ with table output" {
  run steampipe query "select '2024-01-15T10:30:45Z'::timestamptz as tstz"
  assert_output --partial "2024-01-15T10:30:45Z"
}

@test "TIMESTAMPTZ respects session timezone" {
  # Default session timezone is UTC
  run steampipe query "show timezone" --output json
  echo "$output" | jq -e '.rows[0].TimeZone == "UTC"'
  assert_success
}

@test "TIMESTAMPTZ NULL value" {
  run steampipe query "select null::timestamptz as null_tstz" --output json
  echo "$output" | jq -e '.rows[0].null_tstz == null'
  assert_success
}

@test "TIMESTAMP displays without timezone" {
  run steampipe query "select '1984-01-01 12:30:45'::timestamp as ts" --output json
  echo "$output" | jq -e '.rows[0].ts == "1984-01-01 12:30:45"'
  assert_success
}

@test "TIME displays correctly" {
  run steampipe query "select '15:30:45'::time as time_val" --output json
  echo "$output" | jq -e '.rows[0].time_val == "15:30:45"'
  assert_success
}

@test "INTERVAL displays correctly" {
  run steampipe query "select '1 year 2 months 3 days'::interval as interval_val"
  assert_output --partial "1 year 2 mons 3 days"
}

@test "Multiple date/time types together" {
  run steampipe query "select '2024-01-15'::date as d, '2024-01-15 10:30:00'::timestamp as ts, '2024-01-15T10:30:00Z'::timestamptz as tstz" --output json

  # Verify DATE has no time component
  echo "$output" | jq -e '.rows[0].d == "2024-01-15"'
  assert_success

  # Verify TIMESTAMP has time but no timezone
  echo "$output" | jq -e '.rows[0].ts == "2024-01-15 10:30:00"'
  assert_success

  # Verify TIMESTAMPTZ has timezone
  echo "$output" | jq -e '.rows[0].tstz == "2024-01-15T10:30:00Z"'
  assert_success
}

@test "DATE CSV output" {
  run steampipe query "select '1984-01-01'::date as date_val" --output csv
  assert_output --partial "date_val"
  assert_output --partial "1984-01-01"
  refute_output --partial "00:00:00"
}

@test "TIMESTAMPTZ CSV output" {
  run steampipe query "select '1984-01-01T00:00:00Z'::timestamptz as tstz" --output csv
  assert_output --partial "tstz"
  assert_output --partial "1984-01-01T00:00:00Z"
}

@test "DATE line output" {
  run steampipe query "select '1984-01-01'::date as date_val" --output line
  assert_output --partial "date_val"
  assert_output --partial "1984-01-01"
  refute_output --partial "00:00:00"
}

@test "DATE array" {
  run steampipe query "select array['2024-01-01'::date, '2024-12-31'::date] as date_array" --output json
  echo "$output" | jq -e '.rows[0].date_array == "2024-01-01,2024-12-31"'
  assert_success
}

@test "TIMESTAMPTZ edge case - leap year" {
  run steampipe query "select '2024-02-29T23:59:59Z'::timestamptz as leap_tstz" --output json
  echo "$output" | jq -e '.rows[0].leap_tstz == "2024-02-29T23:59:59Z"'
  assert_success
}

@test "TIMESTAMPTZ edge case - year 1" {
  run steampipe query "select '0001-01-01T00:00:00Z'::timestamptz as min_tstz" --output json
  assert_success
}

@test "DATE comparison preserves semantics" {
  # Verify that DATE values can be compared correctly
  run steampipe query "select ('2024-01-15'::date > '2024-01-01'::date) as result" --output json
  echo "$output" | jq -e '.rows[0].result == true'
  assert_success
}

@test "now() returns timestamptz in UTC" {
  run steampipe query "select now()::timestamptz::text as now_val" --output json
  # Should end with Z or +00:00 (UTC)
  echo "$output" | jq -r '.rows[0].now_val' | grep -E '(Z|\+00:00)$'
  assert_success
}

@test "current_date returns date without time" {
  run steampipe query "select current_date as today" --output json
  # Should not contain time component
  echo "$output" | jq -r '.rows[0].today' | grep -v ':'
  assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
