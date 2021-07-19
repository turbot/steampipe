select
    -- Required Columns
    arn as resource,
    case
        when o ->> 'DomainName' not like '%s3.amazonaws.com' then 'skip'
        when o ->> 'DomainName' like '%s3.amazonaws.com'
            and o -> 'S3OriginConfig' ->> 'OriginAccessIdentity' = '' then 'alarm'
        else 'ok'
        end as status,
    case
        when o ->> 'DomainName' not like '%s3.amazonaws.com' then title || ' origin type is not s3.'
        when o ->> 'DomainName' like '%s3.amazonaws.com'
            and o -> 'S3OriginConfig' ->> 'OriginAccessIdentity' = '' then title || ' origin access identity not configured.'
        else title || ' origin access identity configured.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_cloudfront_distribution,
    jsonb_array_elements(origins) as o;
