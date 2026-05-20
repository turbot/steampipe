load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# Regression test for the PG18 FDW expression_tree_walker bug:
# PG18 dropped the T_RestrictInfo case from expression_tree_walker
# (src/backend/nodes/nodeFuncs.c). When the FDW's
# isAttrInRestrictInfo / extractColumns called pull_var_clause on a
# restrictinfo->clause subtree, and that subtree contained a nested
# RestrictInfo (orclause cache / certain equivalence-class derived
# join predicates), the walker hit the unrecognized tag and elogged
#
#   ERROR: unrecognized node type: 318 (SQLSTATE XX000)
#
# The bug fired on any multi-table JOIN across two foreign tables on
# PG18. Fixed in steampipe-postgres-fdw by adding a PG18-guarded
# strip_restrictinfo_mutator pre-pass before pull_var_clause; the
# tests below exercise the planner path that surfaced the bug.

@test "multi-table foreign-table left join planning succeeds (PG18 RestrictInfo regression gate)" {
  # Self-join across the same chaos foreign table. Pre-fix on PG18
  # this errored with `unrecognized node type: 318` during planning.
  # Post-fix (and on PG14, where the walker has the T_RestrictInfo
  # case) this returns rows.
  run steampipe query --output json "select a.id, b.id as b_id from chaos.chaos_all_column_types a left join chaos.chaos_all_column_types b on b.id = a.id order by a.id::int limit 3"
  assert_success
  refute_output --partial 'unrecognized node type'
  refute_output --partial 'SQLSTATE XX000'
}

@test "multi-table foreign-table inner join with grouped aggregate (PG18 RestrictInfo regression gate)" {
  # GROUP BY + COUNT() over a foreign-table join — the original
  # failing shape the customer-equivalent AWS query surfaced.
  run steampipe query --output json "select a.id, count(b.id) as n from chaos.chaos_all_column_types a left join chaos.chaos_all_column_types b on b.id = a.id group by a.id order by a.id::int limit 3"
  assert_success
  refute_output --partial 'unrecognized node type'
  refute_output --partial 'SQLSTATE XX000'
}

function setup() {
  # The same skip-on-arm64-Linux guard the other chaos bats files use
  # (no chaos plugin binary published for that combo).
  sys=$(uname -sm)
  if [[ "$sys" == "Linux aarch64" ]]; then
    skip
  fi
}
