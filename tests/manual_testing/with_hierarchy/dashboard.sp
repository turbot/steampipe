dashboard "with_hiearchy" {
  with "dashboard_with" {
    sql = <<-EOQ
          select 'dashboard_with'
        EOQ
  }
  title         = "With hierarchy"


  graph {
    with "graph_with" {
      sql = <<-EOQ
          select 'graph_with'
        EOQ
    }
    title = "Relationships"
    width = 12
    type  = "graph"


    node "n1" {
      with "node_with1" {
        sql = <<-EOQ
          select 'node_with1'
        EOQ
      }

      sql = <<-EOQ
    select
      $1 as id,
      $2 as title
EOQ

      args = [with.node_with1.rows[0], with.graph_with.rows[0]]

    }
    node "n2" {
      sql = <<-EOQ
    select
      $1 as id,
      $1 as title
EOQ

      args = [with.dashboard_with.rows[0], with.graph_with.rows[0]]
    }
    edge "n1_n2" {
      sql = <<-EOQ
    select
      $1 as from_id,
      $2 as to_id

EOQ

      args = [with.dashboard_with.rows[0], with.node_with1.rows[0]]
    }


  }
}


