dashboard "name_graph" {
  title         = "named graph with base and args"
  input "bucket_arn" {
    title = "Select a bucket:"
    query = query.s3_bucket_input
    width = 4
  }
  with "bucket_policy" {
    sql = <<-EOQ
      select
        policy_std
      from
        aws_s3_bucket
      where
        arn = $1;
    EOQ
    args = [self.input.bucket_arn.value]
  }
  graph {
    base = graph.iam_policy_structure
    args = {
      policy_std = with.bucket_policy.rows[0].policy_std
    }
  }
}

query "s3_bucket_input" {
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

graph "iam_policy_structure" {
  title = "IAM Policy"
  param "policy_std" {}
  node {
    base = node.iam_policy_statement_action_notaction
    args = {
      iam_policy_std = param.policy_std
    }
  }
}

node "iam_policy_statement_action_notaction" {
  sql = <<-EOQ
    select
      concat('action:', action) as id,
      action as title
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_array_elements_text(coalesce(t.stmt -> 'Action','[]'::jsonb) || coalesce(t.stmt -> 'NotAction','[]'::jsonb)) as action
  EOQ

  param "iam_policy_std" {}
}

