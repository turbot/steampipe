dashboard "testing_with_blocks" {
  title = "Testing with blocks"

  container {
    graph "with_testing" {
      title = "Relationships"
      width = 12
      type  = "graph"

      with "limit_value" {
        sql = <<-EOQ
          select 1 as limit_value
        EOQ
      }
      with "distinct_limit_value" {
        sql = <<-EOQ
          select 1 as distinct_limit_value
        EOQ
      }

      nodes = [
        node.chaos_cache_check_1,
        node.chaos_cache_check_2,
      ]

      edges = [
        edge.chaos_cache_check_1,
      ]
    }
  }
}

node "chaos_cache_check_1" {
  sql = <<-EOQ
    select 1 as node_chaos_cache_check_1
  EOQ
}

node "chaos_cache_check_2" {
  sql = <<-EOQ
    select 2 as node_chaos_cache_check_2
  EOQ
}

edge "chaos_cache_check_1" {
  sql = <<-EOQ
    select 1 as edge_chaos_cache_check_1
  EOQ
}
