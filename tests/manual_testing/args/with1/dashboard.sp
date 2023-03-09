
dashboard "bug_column_does_not_exist" {
  title         = "column does not exist"


  input "policy_arn" {
    title = "Select a policy:"
    query = query.test1_aws_iam_policy_input
    width = 4
  }


  container {

    graph {
      title     = "Relationships"
      type      = "graph"
      direction = "left_right" //"TD"

      with "attached_users" {
        sql = <<-EOQ
          select
            u.arn as user_arn
            --,policy_arn
          from
            aws_iam_user as u,
            jsonb_array_elements_text(attached_policy_arns) as policy_arn
          where
            policy_arn = $1;
            --policy_arn = 'arn:aws:iam::aws:policy/AdministratorAccess'
        EOQ

        param policy_arn {
          // commented out becuase input not working here yet..
          // default = self.input.policy_arn.value
          default = "arn:aws:iam::aws:policy/AdministratorAccess"
        }

      }

      with "attached_roles" {
        sql = <<-EOQ
          select
            arn as role_arn
          from
            aws_iam_role,
            jsonb_array_elements_text(attached_policy_arns) as policy_arn
          where
            policy_arn = $1;
        EOQ

        #args = [self.input.policy_arn.value]
        #args = ["arn:aws:iam::aws:policy/AdministratorAccess"]

        param policy_arn {
          //default = self.input.policy_arn.value
          default = "arn:aws:iam::aws:policy/AdministratorAccess"
        }
      }


      nodes = [
        node.test1_aws_iam_policy_node,
        node.test1_aws_iam_user_nodes,
      ]

      edges = [
        edge.test1_aws_iam_policy_from_iam_user_edges,
      ]

      args = {
        policy_arn  = "arn:aws:iam::aws:policy/AdministratorAccess" //self.input.policy_arn.value

        //// works if you hardcode the list
        policy_arns  = ["arn:aws:iam::aws:policy/AdministratorAccess"]

        // this causes  cannot serialize unknown values
        //policy_arns  = [self.input.policy_arn.value]

        user_arns   = [with.attached_users.rows[0].user_arn]
        role_arns   = with.attached_roles.rows[*].role_arn

      }
    }

  }
}

query "test1_aws_iam_policy_input" {
  sql = <<-EOQ
    with policies as (
      select
        title as label,
        arn as value,
        json_build_object(
          'account_id', account_id
        ) as tags
      from
        aws_iam_policy
      where
        not is_aws_managed

      union all select
        distinct on (arn)
        title as label,
        arn as value,
        json_build_object(
          'account_id', 'AWS Managed'
        ) as tags
      from
        aws_iam_policy
      where
        is_aws_managed
    )
    select
      *
    from
      policies
    order by
      label;
  EOQ
}



node "test1_aws_iam_policy_node" {
  sql = <<-EOQ
    select
      distinct on (arn)
      arn as id,
      name as title,
      jsonb_build_object(
        'ARN', arn,
        'AWS Managed', is_aws_managed::text,
        'Attached', is_attached::text,
        'Create Date', create_date,
        'Account ID', account_id
      ) as properties
    from
      aws_iam_policy
    where
      arn = $1;
  EOQ

  param "policy_arn" {}
}


node "test1_aws_iam_user_nodes" {

  sql = <<-EOQ
    select
      arn as id,
      name as title,
      jsonb_build_object(
        'ARN', arn,
        'Path', path,
        'Create Date', create_date,
        'MFA Enabled', mfa_enabled::text,
        'Account ID', account_id
      ) as properties
    from
      aws_iam_user
    where
      arn = any($1::text[]);
  EOQ

  param "user_arns" {}
}



edge "test1_aws_iam_policy_from_iam_user_edges" {
  title = "attaches"

  sql = <<-EOQ
   select
      policy_arns as to_id,
      user_arns as from_id
    from
      unnest($1::text[]) as policy_arns,
      unnest($2::text[]) as user_arns
  EOQ

  param "policy_arns" {}
  param "user_arns" {}

}
