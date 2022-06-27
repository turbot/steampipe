dashboard dashboard_pos_args {
  title = "dashboard with positional arguments"

  input "user" {
    title = "AWS IAM User"
    sql   = query.query1.sql
    width = 4
  }

  table {
    query = query.query2
    args  = [ self.input.user.value, "foo" ]

    column "depth" {
      display = "none"
    }
  }
}
