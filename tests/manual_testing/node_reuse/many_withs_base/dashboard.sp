dashboard "many_withs_base" {
  title = "Many Withs Base"
  with "n1" {
    query = query.dashboard_with
  }
  graph "foo"{
    base = graph.g1
  }
#
#  graph "bar"{
#    node "n1" {
#      sql = <<-EOQ
#    select
#      $1 as id,
#      $1 as title
#EOQ
#      args = [ with.n1.rows[0]]
#    }
#  }
}


graph "g1"{
  with "n1" {
    query = query.graph_with
  }
  node "n1" {
    sql = <<-EOQ
    select
      $1 as id,
      $1 as title
EOQ
    args = [ with.n1.rows[0]]
  }
    args = [ with.n1.rows[0]]
}



query "graph_with"{
  sql = <<-EOQ
          select 'n1_graph'
        EOQ
}

query "dashboard_with"{
  sql = <<-EOQ
          select 'n1_dashboard'
        EOQ
}
