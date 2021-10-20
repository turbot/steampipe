query "query_1"{
    title ="query_1"
    description = "Simple query 1"
    sql = "select 'ok' as status, 'steampipe' as resource, 'acceptance tests' as reason"
}

query "query_2"{
    title ="query_2"
    description = "Simple query 2"
    sql = "select 'alarm' as status, 'turbot' as resource, 'integration tests' as reason"
}
