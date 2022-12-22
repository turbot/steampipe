dashboard "param_runtime_dep" {
  title         = "param_runtime_dep"

  container {
    graph {
      title = "Relationships"
      width = 12
      type  = "graph"
      param "subnet_ids" {
        default =   with.n1.rows[*]
      }
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
      with "n3" {
        sql = <<-EOQ
          select 'n2'
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
}

