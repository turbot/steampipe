query "control_with_blank_dimension"{
    title ="query_1"
    description = "Simple query 1"
    sql = <<-EOQ
      select 'ok' as status, 'resource 1' as resource, 'reason 1' as reason, 'nb1' as dimension1, '' as dimension2, 'nb3' as dimension3
      UNION ALL
      select 'ok' as status, 'resource 2' as resource, 'reason 2' as reason, 'nb1' as dimension1, 'nb2' as dimension2, 'nb3' as dimension3
    EOQ
}
