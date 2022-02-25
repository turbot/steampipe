select
    timestamp,
    average,
    instance_id
from
    aws_ec2_instance_metric_cpu_utilization_hourly
order by
    timestamp,
    average;