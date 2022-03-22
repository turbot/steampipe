select
    -- Required Columns
    'arn:' || partition || ':apigateway:' || region || '::/apis/' || rest_api_id || '/stages/' || name as resource,
    case
        when method_settings -> '*/*' ->> 'CachingEnabled' = 'true'
            and method_settings -> '*/*' ->> 'CacheDataEncrypted' = 'true' then 'ok'
        else 'alarm'
        end as status,
    case
        when method_settings -> '*/*' ->> 'CachingEnabled' = 'true'
            and method_settings -> '*/*' ->> 'CacheDataEncrypted' = 'true'
            then title || ' API cache and encryption enabled.'
        else title || ' API cache and encryption not enabled.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_api_gateway_stage;