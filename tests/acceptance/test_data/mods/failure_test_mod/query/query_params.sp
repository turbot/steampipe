query "query_with_no_param_defs"{
  description = "query with no parameter definitions"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
}

query "query_with_param_defs_no_defaults"{
  description = "query with parameter definitions but no defaults"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
    description = "First parameter"
  }
  param "p2"{
    description = "Second parameter"
  }
  param "p3"{
    description = "Third parameter"
  }
}