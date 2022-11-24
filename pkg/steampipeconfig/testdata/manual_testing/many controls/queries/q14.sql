select
    -- Required Columns
    arn as resource,
    case
        when default_root_object = '' then 'alarm'
        else 'ok'
        end as status,
    case
        when default_root_object = '' then title || ' default root object not configured.'
        else title || ' default root object configured.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_cloudfront_distribution;