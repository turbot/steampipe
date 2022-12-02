
dashboard "with_no_results" {

  container {

    table {
      title     = "Relationships"
      type      = "graph"
      direction = "TD"

      with "no_results" {
        sql = "select * from  aws_iam_user where arn = 'noooo'"
      }

      query = query.array_arg
      args = {
        arns =  with.no_results.rows[*].arn
      }
    }
  }

}
