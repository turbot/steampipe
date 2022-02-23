select
    period_start,
    service,
    unblended_cost_amount
from aws_cost_by_service_monthly
order by period_start asc, service asc
limit 50