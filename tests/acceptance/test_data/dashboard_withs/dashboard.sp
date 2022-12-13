dashboard "testing_with_blocks" {
  title = "Testing with blocks in graphs"

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

  graph "with_testing" {
    title = "Relationships"
    width = 12
    type  = "graph"

    node "chaos_cache_check_1" {
      sql = <<-EOQ
        select 1 as node_chaos_cache_check_1
      EOQ
    }

    node "chaos_cache_check_2" {
      base = node.chaos_cache_check_top
    }

    edge "chaos_cache_check_1" {
      sql = <<-EOQ
        select 1 as edge_chaos_cache_check_1
      EOQ
    }
  }
}

node "chaos_cache_check_top" {
  sql = <<-EOQ
    select 1 as node_chaos_cache_check_top
  EOQ
}