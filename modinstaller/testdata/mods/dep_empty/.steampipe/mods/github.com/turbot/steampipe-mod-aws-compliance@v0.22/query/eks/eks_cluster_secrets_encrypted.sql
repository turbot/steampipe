with eks_secrets_encrypted as (
  select
    distinct arn as arn
  from
    aws_eks_cluster,
    jsonb_array_elements(encryption_config) as e
  where
    e -> 'Resources'  @> '["secrets"]'
)
select
  -- Required Columns
  a.arn as resource,
  case
    when encryption_config is null then 'alarm'
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when encryption_config is null then a.title || ' encryption not enabled.'
    when b.arn is not null then a.title || ' encrypted with EKS secrets.'
    else a.title || ' not encrypted with EKS secrets.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_eks_cluster as a
  left join eks_secrets_encrypted as b on a.arn = b.arn;