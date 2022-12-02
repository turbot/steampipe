dashboard "bug5_multiple_inputs" {
  title         = "bug5: multiple inputs"
  input "bucket_arn" {
    title = "Select a bucket:"
    sql   = query.bug5_s3_bucket_input.sql
    width = 4
  }
  graph {
    title     = "Relationships 1"
    type      = "graph"
    //direction = "left_right" 
    with "policy_std" {
      sql = <<-EOQ
        select
          policy_std
        from
          aws_s3_bucket
        where
          arn = $1
      EOQ
      args = [self.input.bucket_arn.value]
    }
    nodes = [
      node.bug5_me_node,
      node.bug5_iam_policy_statement_nodes,
    ]
    edges = [
      edge.bug5_bucket_policy_statement_edges,
    ]
    args = {
      arn = self.input.bucket_arn.value
      policy_std  = with.policy_std.rows[0].policy_std
      bucket_arns = [self.input.bucket_arn.value]
    }
  }
  table {
    # sql = <<-EOQ
    sql = <<-EOQ
      select
        concat('statement:', i) as id,
        coalesce (
          t.stmt ->> 'Sid',
          concat('[', i::text, ']')
          ) as title
      from
        aws_s3_bucket,
        jsonb_array_elements(policy_std ->  'Statement') with ordinality as t(stmt,i)
      where
        arn = $1
    EOQ
    args = [self.input.bucket_arn.value]
  }
  //***************
  input "lambda_arn" {
    title = "Select a lambda function:"
    sql   = query.bug5_lambda_function_input.sql
    width = 4
  }
  graph {
    title     = "Relationships 2"
    type      = "graph"
    //direction = "left_right" 
    with "policy_std" {
      sql = <<-EOQ
        select
          policy_std
        from
          aws_lambda_function
        where
          arn = $1
      EOQ
      args = [self.input.lambda_arn.value]
    }
    nodes = [
      node.bug5_me_node,
      node.bug5_iam_policy_statement_nodes,
    ]
    edges = [
      edge.bug5_lambda_function_policy_statement_edges,
    ]
    args = {
      arn = self.input.lambda_arn.value
      policy_std  = with.policy_std.rows[0].policy_std
      lambda_arns = [self.input.lambda_arn.value]
    }
  }
  table {
    sql = <<-EOQ
      select
        concat('statement:', i) as id,
        coalesce (
          t.stmt ->> 'Sid',
          concat('[', i::text, ']')
          ) as title
      from
        aws_lambda_function,
        jsonb_array_elements(policy_std ->  'Statement') with ordinality as t(stmt,i)
      where
        arn = $1
    EOQ

    args = [self.input.lambda_arn.value]
  }

}
query "bug5_s3_bucket_input" {
  sql = <<-EOQ
    select
      title as label,
      arn as value,
      json_build_object(
        'account_id', account_id,
        'region', region
      ) as tags
    from
      aws_s3_bucket
    order by
      title;
  EOQ
}
query "bug5_lambda_function_input" {
  sql = <<-EOQ
    select
      title as label,
      arn as value,
      json_build_object(
        'account_id', account_id,
        'region', region
      ) as tags
    from
      aws_lambda_function
    order by
      title;
  EOQ
}
node "bug5_me_node" {
  //category = category.aws_iam_policy
  sql = <<-EOQ
    select
      $1 as id,
      $1 as title
  EOQ
  param "arn" {}
}
node "bug5_iam_policy_statement_nodes" {
  //category = category.aws_iam_policy_statement
  sql = <<-EOQ
    select
      concat('statement:', i) as id,
      coalesce (
        t.stmt ->> 'Sid',
        concat('[', i::text, ']')
        ) as title
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i)
  EOQ
  param "policy_std" {}
}
edge "bug5_bucket_policy_statement_edges" {
  title = "statement"
  sql = <<-EOQ
    
    select
      distinct on (arn,i)
      arn as from_id,
      concat('statement:', i) as to_id
    from
      aws_s3_bucket,
      jsonb_array_elements(policy_std -> 'Statement') with ordinality as t(stmt,i)
    where
      arn = any($1)
  EOQ
  param "bucket_arns" {}
}
edge "bug5_lambda_function_policy_statement_edges" {
  title = "statement"
  sql = <<-EOQ
    
    select
      distinct on (arn,i)
      arn as from_id,
      concat('statement:', i) as to_id
    from
      aws_lambda_function,
      jsonb_array_elements(policy_std -> 'Statement') with ordinality as t(stmt,i)
    where
      arn = any($1)
  EOQ
  param "lambda_arns" {}
}