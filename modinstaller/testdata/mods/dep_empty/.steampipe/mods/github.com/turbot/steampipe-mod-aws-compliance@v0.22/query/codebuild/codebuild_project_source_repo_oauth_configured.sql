select
  -- Required Columns
  p.arn as resource,
  case
    when p.source ->> 'Type' not in ('GITHUB', 'BITBUCKET') then 'skip'
    when c.auth_type = 'OAUTH' then 'ok'
    else 'alarm'
  end as status,
  case
    when p.source ->> 'Type' = 'NO_SOURCE' then p.title || ' doesn''t have input source code.'
    when p.source ->> 'Type' not in ('GITHUB', 'BITBUCKET') then p.title || ' source code isn''t in GitHub/Bitbucket repository.'
    when c.auth_type = 'OAUTH' then p.title || ' using OAuth to connect source repository.'
    else p.title || ' not using OAuth to connect source repository.'
  end as reason,
  -- Additional Dimensions
  p.region,
  p.account_id
from
  aws_codebuild_project as p
  left join aws_codebuild_source_credential as c on (p.region = c.region and p.source ->> 'Type' = c.server_type);