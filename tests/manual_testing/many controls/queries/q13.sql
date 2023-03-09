with data as (
    select
        distinct arn
    from
        aws_cloudfront_distribution,
        jsonb_array_elements(
                case jsonb_typeof(cache_behaviors -> 'Items')
                    when 'array' then (cache_behaviors -> 'Items')
                    else null end
            ) as cb
    where
                cb -> 'ViewerProtocolPolicy' = '"allow-all"'
)
select
    -- Required Columns
    b.arn as resource,
    case
        when d.arn is not null or (default_cache_behavior ->> 'ViewerProtocolPolicy' = 'allow-all') then 'alarm'
        else 'ok'
        end as status,
    case
        when d.arn is not null or (default_cache_behavior ->> 'ViewerProtocolPolicy' = 'allow-all') then title || ' data not encrypted in transit.'
        else title || ' data encrypted in transit.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_cloudfront_distribution as b
        left join data as d on b.arn = d.arn;

