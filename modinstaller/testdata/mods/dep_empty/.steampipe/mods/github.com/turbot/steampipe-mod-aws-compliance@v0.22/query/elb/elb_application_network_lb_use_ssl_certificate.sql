 with listeners_without_certificate as (
  select
    load_balancer_arn,
    count(*) as count
  from
    aws_ec2_load_balancer_listener
  where arn not in
    ( select arn from aws_ec2_load_balancer_listener, jsonb_array_elements(certificates) as c
      where c ->> 'CertificateArn' like 'arn:aws:acm%' )
  group by load_balancer_arn
),
all_application_network_load_balacer as (
  select
    arn,
    account_id,
    region,
    title
  from
    aws_ec2_application_load_balancer
  union
  select
    arn,
    account_id,
    region,
    title
  from
    aws_ec2_network_load_balancer
)
select
  -- Required Columns
  a.arn as resource,
  case
    when b.load_balancer_arn is null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.load_balancer_arn is null then a.title || ' uses certificates provided by ACM.'
    else a.title || ' has ' || b.count || ' listeners which do not use certificates provided by ACM.'
  end as reason,
  -- Additional Dimensions
  a.region,
  a.account_id
from
  all_application_network_load_balacer as a
  left join listeners_without_certificate as b on a.arn = b.load_balancer_arn;