dashboard "bug_passing_json" {
  title         = "Bug: Passing JSON"


    graph {
      title     = "Relationships"
      type      = "graph"
      direction = "left_right" //"TD"

      with "policy_std" {
        sql = <<-EOQ
          select
            policy_std
          from
            aws_iam_policy
          where
            arn = $1
          limit 1;  -- aws managed policies will appear once for each connection in the aggregator, but we only need one...
        EOQ

        #args = [self.input.policy_arn.value]
        #args = ["arn:aws:iam::aws:policy/AdministratorAccess"]

        param policy_arn {
          //default = self.input.policy_arn.value
          default = "arn:aws:iam::aws:policy/AdministratorAccess"
        }
      }


      nodes = [
        //node.aws_iam_policy_nodes,
        node.test4_aws_iam_policy_statement_nodes,
      ]

      edges = [
      ]

      args = {
        policy_std    = with.policy_std.rows[0].policy_std
        //policy_std    = with.policy_std.rows[*].policy_std

      }
    }
}



node "test4_aws_iam_policy_statement_nodes" {

  sql = <<-EOQ

    select
      concat('statement:', i) as id,
      coalesce (
        t.stmt ->> 'Sid',
        concat('[', i::text, ']')
        ) as title
    from
      (select $1) as p,
      jsonb_array_elements(to_jsonb(p) -> 'jsonb' -> 'Statement') with ordinality as t(stmt,i)

  EOQ

  param "policy_std" {}
}