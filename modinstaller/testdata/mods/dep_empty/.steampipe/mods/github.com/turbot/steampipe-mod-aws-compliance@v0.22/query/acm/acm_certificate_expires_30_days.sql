select
  -- Required Columns
  certificate_arn as resource,
  case
    when renewal_eligibility = 'INELIGIBLE' then 'skip'
    when not_after <= (current_date - interval '30' day) then 'ok'
    else 'alarm'
  end as status,
  case
    when renewal_eligibility = 'INELIGIBLE' then title || ' not eligible for renewal.'
    else title || ' expires ' || to_char(not_after, 'DD-Mon-YYYY') ||
    ' (' || extract(day from not_after - current_timestamp) || ' days).'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_acm_certificate;
