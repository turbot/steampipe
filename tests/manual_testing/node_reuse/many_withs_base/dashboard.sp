dashboard "many_withs_base" {
  title = "Many Withs Base"
  graph {
    base = graph.g1
  }
}


graph "g1"{
  with "n1" {
    sql = <<-EOQ
          select 'n1'
        EOQ
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




