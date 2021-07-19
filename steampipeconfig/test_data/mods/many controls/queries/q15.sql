select
    -- Required Columns
    arn as resource,
    case
        when origin_groups ->> 'Items' is not null then 'ok'
        else 'alarm'
        end as status,
    case
        when origin_groups ->> 'Items' is not null then title || ' origin group is configured.'
        else title || ' origin group not configured.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_cloudfront_distribution;