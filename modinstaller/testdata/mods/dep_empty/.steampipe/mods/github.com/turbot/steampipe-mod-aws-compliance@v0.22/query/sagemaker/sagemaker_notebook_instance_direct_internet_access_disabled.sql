select
  -- Required Columns
  arn as resource,
  case
    when direct_internet_access = 'Enabled' then 'alarm'
    else 'ok'
  end status,
  case
    when direct_internet_access = 'Enabled' then title || ' direct internet access enabled.'
    else title || ' direct internet access disabled.'
  end reason,
  -- Additional Dimentions
  region,
  account_id
from
  aws_sagemaker_notebook_instance;