select
    -- Required Columns
    autoscaling_group_arn as resource,
    case
        when load_balancer_names is null and target_group_arns is null then 'alarm'
        when health_check_type != 'ELB' then 'alarm'
        else 'ok'
        end as status,
    case
        when load_balancer_names is null and target_group_arns is null then title || ' not associated with a load balancer.'
        when health_check_type != 'ELB' then title || ' does not use ELB health check.'
        else title || ' uses ELB health check.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_ec2_autoscaling_group;
