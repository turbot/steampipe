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



//**  The Graph....

graph "iam_policy_structure" {
  title = "IAM Policy"

  param "policy_std" {}

  # node {
  #   base = node.iam_policy_statement
  #   args = {
  #     iam_policy_std = param.policy_std
  #   }
  # }

  node {
    base = node.iam_policy_statement_action_notaction
    args = {
      iam_policy_std = param.policy_std
    }
  }

  node {
    base = node.iam_policy_statement_condition
    args = {
      iam_policy_std = param.policy_std
    }
  }

  node {
    base = node.iam_policy_statement_condition_key
    args = {
      iam_policy_std = param.policy_std
    }
  }

  node {
    base = node.iam_policy_statement_condition_key_value
    args = {
      iam_policy_std = param.policy_std
    }
  }

  node {
    base = node.iam_policy_statement_resource_notresource
    args = {
      iam_policy_std = param.policy_std
    }
  }


  # edge {
  #   base = edge.iam_policy_statement
  #   args = {
  #     iam_policy_arns = [self.input.policy_arn.value]
  #   }
  # }

  edge {
    base = edge.iam_policy_statement_action
    args = {
      iam_policy_std = param.policy_std
    }
  }

  edge {
    base = edge.iam_policy_statement_condition
    args = {
      iam_policy_std = param.policy_std
    }
  }

  edge {
    base = edge.iam_policy_statement_condition_key
    args = {
      iam_policy_std = param.policy_std
    }
  }

  edge {
    base = edge.iam_policy_statement_condition_key_value
    args = {
      iam_policy_std = param.policy_std
    }
  }

  edge {
    base = edge.iam_policy_statement_notaction
    args = {
      iam_policy_std = param.policy_std
    }
  }

  edge {
    base = edge.iam_policy_statement_notresource
    args = {
      iam_policy_std = param.policy_std
    }
  }

  edge {
    base = edge.iam_policy_statement_resource
    args = {
      iam_policy_std = param.policy_std
    }
  }
}



// nodes


node "iam_policy_statement" {
  category = category.iam_policy_statement

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

  param "iam_policy_std" {}
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

node "iam_policy_statement_condition" {
  category = category.iam_policy_condition

  sql = <<-EOQ
    select
      condition.key as title,
      concat('statement:', i, ':condition:', condition.key  ) as id,
      condition.value as properties
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_each(t.stmt -> 'Condition') as condition
    where
      stmt -> 'Condition' <> 'null'
  EOQ

  param "iam_policy_std" {}
}

node "iam_policy_statement_condition_key" {
  category = category.iam_policy_condition_key

  sql = <<-EOQ
    select
      condition_key.key as title,
      concat('statement:', i, ':condition:', condition.key, ':', condition_key.key  ) as id,
      condition_key.value as properties
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_each(t.stmt -> 'Condition') as condition,
      jsonb_each(condition.value) as condition_key
    where
      stmt -> 'Condition' <> 'null'
  EOQ

  param "iam_policy_std" {}
}

node "iam_policy_statement_condition_key_value" {
  category = category.iam_policy_condition_value

  sql = <<-EOQ
    select
      condition_value as title,
      concat('statement:', i, ':condition:', condition.key, ':', condition_key.key, ':', condition_value  ) as id
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_each(t.stmt -> 'Condition') as condition,
      jsonb_each(condition.value) as condition_key,
      jsonb_array_elements_text(condition_key.value) as condition_value
    where
      stmt -> 'Condition' <> 'null'
  EOQ

  param "iam_policy_std" {}
}

node "iam_policy_statement_resource_notresource" {
  category = category.iam_policy_resource

  sql = <<-EOQ
    select
      resource as id,
      resource as title
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_array_elements_text(coalesce(t.stmt -> 'Action','[]'::jsonb) || coalesce(t.stmt -> 'NotAction','[]'::jsonb)) as action,
      jsonb_array_elements_text(coalesce(t.stmt -> 'Resource','[]'::jsonb) || coalesce(t.stmt -> 'NotResource','[]'::jsonb)) as resource
  EOQ

  param "iam_policy_std" {}
}


// edges

edge "iam_policy_statement_action" {
  //title = "allows"
  sql = <<-EOQ

    select
      --distinct on (p.arn,action)
      concat('action:', action) as to_id,
      concat('statement:', i) as from_id,
      lower(t.stmt ->> 'Effect') as title,
      lower(t.stmt ->> 'Effect') as category
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_array_elements_text(t.stmt -> 'Action') as action
  EOQ

  param "iam_policy_std" {}
}

edge "iam_policy_statement_condition" {
  title = "condition"
  sql   = <<-EOQ

    select
      concat('statement:', i, ':condition:', condition.key) as to_id,
      concat('statement:', i) as from_id
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_each(t.stmt -> 'Condition') as condition
    where
      stmt -> 'Condition' <> 'null'
  EOQ

  param "iam_policy_std" {}
}

edge "iam_policy_statement_condition_key" {
  title = "all of"
  sql   = <<-EOQ
    select
      concat('statement:', i, ':condition:', condition.key, ':', condition_key.key  ) as to_id,
      concat('statement:', i, ':condition:', condition.key) as from_id
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_each(t.stmt -> 'Condition') as condition,
      jsonb_each(condition.value) as condition_key
    where
      stmt -> 'Condition' <> 'null'
  EOQ

  param "iam_policy_std" {}
}

edge "iam_policy_statement_condition_key_value" {
  title = "any of"
  sql   = <<-EOQ
    select
      concat('statement:', i, ':condition:', condition.key, ':', condition_key.key, ':', condition_value  ) as to_id,
      concat('statement:', i, ':condition:', condition.key, ':', condition_key.key  ) as from_id
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_each(t.stmt -> 'Condition') as condition,
      jsonb_each(condition.value) as condition_key,
      jsonb_array_elements_text(condition_key.value) as condition_value
    where
      stmt -> 'Condition' <> 'null'
  EOQ

  param "iam_policy_std" {}
}

edge "iam_policy_statement_notaction" {
  sql = <<-EOQ

    select
      --distinct on (p.arn,notaction)
      concat('action:', notaction) as to_id,
      concat('statement:', i) as from_id,
      concat(lower(t.stmt ->> 'Effect'), ' not action') as title,
      lower(t.stmt ->> 'Effect') as category
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i),
      jsonb_array_elements_text(t.stmt -> 'NotAction') as notaction
  EOQ

  param "iam_policy_std" {}
}

edge "iam_policy_statement_notresource" {
  title = "not resource"

  sql = <<-EOQ
    select
      concat('action:', coalesce(action, notaction)) as from_id,
      notresource as to_id,
      lower(stmt ->> 'Effect') as category
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i)
      left join jsonb_array_elements_text(stmt -> 'Action') as action on true
      left join jsonb_array_elements_text(stmt -> 'NotAction') as notaction on true
      left join jsonb_array_elements_text(stmt -> 'NotResource') as notresource on true
  EOQ

  param "iam_policy_std" {}
}

edge "iam_policy_statement_resource" {
  title = "resource"

  sql = <<-EOQ
    select
      concat('action:', coalesce(action, notaction)) as from_id,
      resource as to_id,
      lower(stmt ->> 'Effect') as category
    from
      jsonb_array_elements(($1 :: jsonb) ->  'Statement') with ordinality as t(stmt,i)
      left join jsonb_array_elements_text(stmt -> 'Action') as action on true
      left join jsonb_array_elements_text(stmt -> 'NotAction') as notaction on true
      left join jsonb_array_elements_text(stmt -> 'Resource') as resource on true
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
