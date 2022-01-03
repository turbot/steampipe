(
  select
    -- Required Columns
    arn as resource,
    case
      when load_balancer_attributes @> '[{"Key": "access_logs.s3.enabled", "Value": "true"}]' then 'ok'
      else 'alarm'
    end as status,
    case
      when load_balancer_attributes @> '[{"Key": "access_logs.s3.enabled", "Value": "true"}]' then title || ' logging enabled.'
      else title || ' logging disabled.'
    end as reason,
    -- Additional Dimensions
    region,
    account_id
  from
    aws_ec2_application_load_balancer
)
union
(
  select
    -- Required Columns
    'arn:' || partition || ':elasticloadbalancing:' || region || ':' || account_id || ':loadbalancer/' || title as resource,
    case
      when access_log_enabled = 'true' then 'ok'
      else 'alarm'
    end as status,
    case
      when access_log_enabled = 'true' then title || ' logging enabled.'
      else title || ' logging disabled.'
    end as reason,
    -- Additional Dimensions
    region,
    account_id
  from
    aws_ec2_classic_load_balancer
);