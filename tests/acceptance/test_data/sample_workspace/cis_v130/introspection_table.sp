benchmark "introspection_table_features" {
  title = "Benchmark to test the introspection table features in steampipe"
  children = [
    control.ensure_mod_name
  ]
}

query "steampipe_query" {
  title = "steampipe query"
  description = "foo"
  sql = "select * from steampipe_query"
}

control "ensure_mod_name" {
  title = "Control to test that the mod name for the resource is as required)"
  query = query.steampipe_query
}