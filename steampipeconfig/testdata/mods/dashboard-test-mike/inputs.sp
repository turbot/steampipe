#variable "aws_iam_user_arn" {
#  type    = string
#  default = "arn:aws:iam::876515858155:user/mike"
#}

#query "aws_iam_user_analysis" {
#  sql = <<EOQ
#with analysis as (
#  select
#    null as parent,
#    arn as id,
#    title as name,
#    0 as depth,
#    'aws_iam_user' as category
#  from
#    aws_iam_user
#  where
#    arn = $1
#  union
#  select
#    u.arn as parent,
#    ak.access_key_id as id,
#    ak.title as name,
#    1 as depth,
#    'aws_iam_access_key' as category
#  from
#    aws_iam_access_key ak
#    inner join aws_iam_user u on ak.user_name = u.name
#  where
#    u.arn = $1
#  union
#  select
#    u.arn as parent,
#    g.arn as id,
#    g.title as name,
#    1 as depth,
#    'aws_iam_group' as category
#  from
#    aws_iam_user u,
#    jsonb_array_elements(u.groups) as user_groups
#    inner join aws_iam_group g on g.arn = user_groups ->> 'Arn'
#  where
#    u.arn = $1
#  union
#  select
#    g.arn as parent,
#    p.arn as id,
#    p.title as name,
#    2 as depth,
#    'aws_iam_policy' as category
#  from
#    aws_iam_user as u,
#    aws_iam_policy as p,
#    jsonb_array_elements(u.groups) as user_groups
#    inner join aws_iam_group g on g.arn = user_groups ->> 'Arn'
#  where
#    g.attached_policy_arns :: jsonb ? p.arn
#    and u.arn = $1
#  union
#  select
#    u.arn as parent,
#    p.arn as id,
#    p.title as name,
#    2 as depth,
#    'aws_iam_policy' as category
#  from
#    aws_iam_user as u,
#    jsonb_array_elements_text(u.attached_policy_arns) as pol_arn,
#    aws_iam_policy as p
#  where
#    u.attached_policy_arns :: jsonb ? p.arn
#    and pol_arn = p.arn
#    and u.arn = $1
#)
#select
#  *
#from
#  analysis
#order by
#  depth,
#  category,
#  id;
#EOQ
#  param "aws_iam_user_arn" {
##    default = var.aws_iam_user_arn
#    default = "arn:aws:iam::876515858155:user/mike"
#  }
#}

# Containers...Image, HCL, Dashboard Inputs / Params

#dashboard
#container

# Leaf nodes...Image, HCL, SQL, Params

#chart
#text
#image
#table
#hierarchy
#counter
#input




#control
#benchmark



table "aws_iam_user_analysis_table" {
  sql = query.aws_iam_user_analysis.sql
  column "depth" {
    display = "none"
  }
}
#
#dashboard "foo" {
#  ...
#}

#query "q1" {
#  param "tags" {}
#  sql = "select * from ..."
#}
#
#container "foo" {
#  param "tags" {}
#
#  chart {
#    query = query.q1
#    args  = {
#      tags = parent.param.tags.value
#    }
#  }
#}
#
#dashboard "has_param" {
#  input "region" {
#    sql = "select name, label from aws_region"
#  }
#
#  chart {
#    query = query.q1
#    args  = {
#      tags = self.input.tags.value
#    }
#  }
#}
#
#dashboard "re_use" {
#  container {
#    base = dashboard.has_param
#  }
#}
#
#dashboard "bar" {
#  input "tags" {}
#
#  container {
#    base = container.foo
#    args = {
#      # pass hard coded arg to container
#      "tags" = ["foo"]
#    }
#  }
#
#  container {
#    base = container.foo
#    args = {
#      # pass hard coded arg to container
#      "tags" = ["bar"]
#    }
#  }
#
#  container {
#    container {
#      base = container.foo
#      args = {
#        # get tags from dashboard input block
#        "tags" = root.input.tags
#      }
#    }
#  }
#}

