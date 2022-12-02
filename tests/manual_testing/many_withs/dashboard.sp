dashboard "many_withs" {
  title         = "Many Withs"

  container {
    graph {
      title = "Relationships"
      width = 12
      type  = "graph"

      with "n1" {
        sql = <<-EOQ
          select 'n1'
        EOQ
      }
      with "n2" {
        sql = <<-EOQ
          select 'n2'
        EOQ
      }
      with "n2" {
        sql = <<-EOQ
          select 'n2'
        EOQ
      }
      with "n3" {
        sql = <<-EOQ
          select 'n2'
        EOQ
      }

      nodes = [
        node.n1,
        node.n2,
      ]

      edges = [
        edge.n1_n2,
      ]

      args = {
        n1 = with.n1.rows[0]
        n2 = with.n2.rows[0]
      }
    }
  }
}

node "n1" {
  sql = <<-EOQ
      select
        $1 as id,
        $1 as title

  EOQ
  param "n1" {}
}
node "n2" {
  sql = <<-EOQ
      select
        $1 as id,
        $1 as title
  EOQ

  param "n2" {}
}

edge "n1_n2" {
  sql = <<-EOQ
      select
        $1 as from_id,
        $2 as to_id

  EOQ

  param "n1" {}
  param "n2" {}
}
