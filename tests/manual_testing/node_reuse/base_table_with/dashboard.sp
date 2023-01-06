dashboard "base_query_with" {
  title = "base_query_with"
  table "foo"{
    base = table.t1
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


table "t1"{
  with "n1" {
    query = query.q1
  }
  sql = "select $1"
  args = [ with.n1.rows[0]]
#  args = ["foo"]
}

query "q1"{
  sql = "select '1'"

}