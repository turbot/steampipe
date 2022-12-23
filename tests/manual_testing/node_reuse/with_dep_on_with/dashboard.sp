dashboard "with_dep_on_with" {
  title         = "With dependent on with"
  with "n1" {
    sql = <<-EOQ
          select 'n1'
        EOQ
  }
  with "n2" {
    sql = <<-EOQ
          select $1
        EOQ
    args = [ with.dependency.rows[0]]
  }
  with "dependency" {
    sql = <<-EOQ

          select 'dependency_with'
        EOQ
  }

  graph {
    title = "Relationships"
    width = 12
    type  = "graph"


    node "n1" {
      sql = <<-EOQ
    select
      $1 as id,
      $1 as title
EOQ
      args = [ with.n1.rows[0]]
    }
    node "n2" {
      sql = <<-EOQ
    select
      $1 as id,
      $1 as title
EOQ

      args = [ with.n2.rows[0]]
    }
    edge "n1_n2" {
      sql = <<-EOQ
    select
      $1 as from_id,
      $2 as to_id
EOQ
      args = [with.n1.rows[0], with.n2.rows[0]]
    }
  }

}
