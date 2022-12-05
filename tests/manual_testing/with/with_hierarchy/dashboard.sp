dashboard "with_hiearchy" {
  with "dashboard_with_node2" {
    sql = <<-EOQ
          select 'dashboard_with_node2'
        EOQ
  }
  title         = "With hierarchy"


  graph {
    with "graph_with_node1" {
      sql = <<-EOQ
          select 'graph_with_node1'
        EOQ
    }
    title = "Relationships"
    width = 12
    type  = "graph"


    node "n1" {
      with "node_with_title" {
        sql = <<-EOQ
          select 'node_with_title'
        EOQ
      }

      sql = <<-EOQ
    select
      $1 as id,
      $2 as title
EOQ

      args = [with.graph_with_node1.rows[0], with.node_with_title.rows[0]]

    }
    node "n2" {
      with "node_with_title" {
        sql = <<-EOQ
          select 'node_with_title'
        EOQ
      }
      sql = <<-EOQ
    select
      $1 as id,
      $2 as title
EOQ

      args = [with.dashboard_with_node2.rows[0], with.node_with_title.rows[0]]
    }
    edge "n1_n2" {
      sql = <<-EOQ
    select
      $1 as from_id,
      $2 as to_id
EOQ
      args = [with.graph_with_node1.rows[0], with.dashboard_with_node2.rows[0]]
    }
  }
}



