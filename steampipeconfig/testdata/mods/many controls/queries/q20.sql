select
    -- Required Columns
    certificate_arn as resource,
    case
        when not_after <= (current_date - interval '30' day) then 'ok'
        else 'alarm'
        end as status,
    title || ' expires ' || to_char(not_after, 'DD-Mon-YYYY') ||
    ' (' || extract(day from not_after - current_timestamp) || ' days).'
                    as reason,
    -- Additional Dimensions
    region,
    account_id
from
    aws_acm_certificate;
