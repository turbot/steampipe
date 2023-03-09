dashboard "query_providers_top_level_require_sql" {
  title = "Query providers at top level that require sql/query block"
  description = "This is a dashboard that validates - Query providers at top level DO NOT need a query/sql block except Control and Query"
}

query "top_query_1" {
  description = "This is a top level query block"
  sql = "select 1 as query"
}


control "top_control_1" {
  description = "This is a top level control block"
  sql = "select 1 as control"
}

control "top_control_2" {
  description = "This is a top level control block"
  query = query.simple_query
}