dashboard "many_withs" {
  input "i1" {
    sql = <<-EOQ
          select arn as label, arn as value from aws_account
        EOQ
    placeholder = "enter a val"
  }


  title         = "Many Withs"
  with "n1" {
   query = query.q1
  }
  with "n2" {
    sql = <<-EOQ
          select $1
        EOQ
    args = [self.input.i1.value]
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

query "q1"{
  sql = <<-EOQ
          select 'n1'
        EOQ
}
