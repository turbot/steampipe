dashboard "many_withs_base" {
  title = "Many Withs Base"
  with "n1" {
    query = query.q2
  }
  graph {
    base = graph.g1
  }

  graph {
    node "n1" {
      sql = <<-EOQ
    select
      $1 as id,
      $1 as title
EOQ
      args = [ with.n1.rows[0]]
    }
  }
}


graph "g1"{
  with "n1" {
    query = query.q1
  }
  node "n1" {
    sql = <<-EOQ
    select
      $1 as id,
      $1 as title
EOQ
    args = [ with.n1.rows[0]]
  }
}



query "q1"{
  sql = <<-EOQ
          select 'n1'
        EOQ
}

query "q2"{
  sql = <<-EOQ
          select 'n1_dashboard'
        EOQ
}
