dashboard dashboard_named_args {
  title = "dashboard with named arguments"

  input "user" {
    title = "AWS IAM User"
    sql   = query.query1.sql
    width = 4
  }

  table {
    query = query.query2
    args  = {
      "iam_user_arn" = self.input.user.value
    }

    column "depth" {
      display = "none"
    }
  }
}