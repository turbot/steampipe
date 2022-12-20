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
      policy_std = with.bucket_policy[0].policy_std
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

//**  The Graph....
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
  category = category.iam_policy_action
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

// categories
category "iam_policy" {
  title = "IAM Policy"
  color = local.iam_color
  href  = "/aws_insights.dashboard.iam_policy_detail?input.policy_arn={{.properties.'ARN' | @uri}}"
  icon  = "rule"
}
category "iam_policy_action" {
  href  = "/aws_insights.dashboard.iam_action_glob_report?input.action_glob={{.title | @uri}}"
  icon  = "electric-bolt"
  color = local.iam_color
  title = "Action"
}
category "iam_policy_condition" {
  icon  = "help"
  color = local.iam_color
  title = "Condition"
}
category "iam_policy_condition_key" {
  icon  = "vpn-key"
  color = local.iam_color
  title = "Condition Key"
}
category "iam_policy_condition_value" {
  icon  = "text:val"
  color = local.iam_color
  title = "Condition Value"
}
category "iam_policy_notaction" {
  icon  = "flash-off"
  color = local.iam_color
  title = "NotAction"
}
category "iam_policy_notresource" {
  icon  = "bookmark-remove"
  color = local.iam_color
  title = "NotResource"
}
category "iam_policy_resource" {
  icon  = "bookmark"
  color = local.iam_color
  title = "Resource"
}
category "iam_policy_statement" {
  icon  = "assignment"
  color = local.iam_color
  title = "Statement"
}
// color
locals {
  analytics_color               = "purple"
  application_integration_color = "deeppink"
  ar_vr_color                   = "deeppink"
  blockchain_color              = "orange"
  business_application_color    = "red"
  compliance_color              = "orange"
  compute_color                 = "orange"
  containers_color              = "orange"
  content_delivery_color        = "purple"
  cost_management_color         = "green"
  database_color                = "blue"
  developer_tools_color          = "blue"
  end_user_computing_color      = "green"
  front_end_web_color           = "red"
  game_tech_color               = "purple"
  iam_color                     = "red"
  iot_color                     = "green"
  management_governance_color   = "pink"
  media_color                   = "orange"
  migration_transfer_color      = "green"
  ml_color                      = "green"
  mobile_color                  = "red"
  networking_color              = "purple"
  quantum_technologies_color    = "orange"
  robotics_color                = "red"
  satellite_color               = "blue"
  security_color                = "red"
  storage_color                 = "green"
}
