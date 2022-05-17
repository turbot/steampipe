with all_stages as (
    select
        name as stage_name,
        'arn:' || partition || ':apigateway:' || region || '::/apis/' || rest_api_id || '/stages/' || name as arn,
        method_settings -> '*/*' ->> 'LoggingLevel' as log_level,
        title,
        region,
        account_id
    from
        aws_api_gateway_stage
    union
    select
        stage_name,
        'arn:' || partition || ':apigateway:' || region || '::/apis/' || api_id || '/stages/' || stage_name as arn,
        default_route_logging_level as log_level,
        title,
        region,
        account_id
    from
        aws_api_gatewayv2_stage
)
select
    -- Required Columns
    arn as resource,
    case
        when log_level is null or log_level = 'OFF' then 'alarm'
        else 'ok'
        end as status,
    case
        when log_level is null or log_level = 'OFF' then title || ' logging not enabled.'
        else title || ' logging enabled.'
        end as reason,
    -- Additional Dimensions
    region,
    account_id
from
    all_stages;