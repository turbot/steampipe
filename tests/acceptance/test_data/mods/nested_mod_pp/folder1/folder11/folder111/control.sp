control "check_1" {
  title         = "Control to verify mod.sp traversal functionality"
  description   = "Control to verify verify mod.sp traversal functionality."
  query         = query.query_1
  severity      = "high"
}

query "query_1"{
  title ="query_1"
  description = "Simple query 1"
  sql = "select 'ok' as status, 'steampipe' as resource, 'acceptance tests' as reason"
}
