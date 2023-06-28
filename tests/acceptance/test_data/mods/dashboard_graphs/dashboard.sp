dashboard "testing_nodes_and_edges" {
  title = "Testing with blocks in graphs"

  graph "node_and_edge_testing" {
    title = "Relationships"
    width = 12
    type  = "graph"

    node "chaos_cache_check_1" {
      sql = <<-EOQ
        select 1 as node_chaos_cache_check_1
      EOQ
    }

    node "chaos_cache_check_2" {
      base = node.chaos_cache_check_top1
    }

    node "chaos_cache_check_3" {
      base = node.chaos_cache_check_top2
    }

    edge "chaos_cache_check_1" {
      sql = <<-EOQ
        select 1 as edge_chaos_cache_check_1
      EOQ
    }

    edge "chaos_cache_check_2" {
      base = edge.chaos_cache_check_top1
    }
  }
}

node "chaos_cache_check_top1" {
  sql = <<-EOQ
    select 1 as node_chaos_cache_check_top
  EOQ
}

node "chaos_cache_check_top2" {
  sql = <<-EOQ
    select 1 as node_chaos_cache_check_top
  EOQ
}

edge "chaos_cache_check_top1" {
  sql = <<-EOQ
    select 1 as edge_chaos_cache_check_2
  EOQ
}
