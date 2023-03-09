dashboard "top_level_control_query_require_sql" {
  title = "Query providers at top level that require sql/query block"
  description = "This is a dashboard that validates - top level controls and queries always require a query/sql block - SHOULD RESULT IN PARSING FAILURE"
}

query "top_query_1" {
  description = "This is a top level query block"
}

query "top_query_2" {
  description = "This is a top level query block"
}

control "top_control_1" {
  description = "This is a top level control block"
}

control "top_control_2" {
  description = "This is a top level control block"
}