(
select
  -- Required Columns
  arn as resource,
  case
      when engine similar to '%(aurora|mysql|mariadb)%' and port = '3306' then 'alarm'
      when engine like '%postgres%' and port = '5432' then 'alarm'
      when engine like 'oracle%' and port = '1521' then 'alarm'
      when engine like 'sqlserver%' and port = '1433' then 'alarm'
    else 'ok'
  end as status,
  case
      when engine similar to '%(aurora|mysql|mariadb)%' and port = '3306' then  title || ' ' ||  engine || ' uses a default port.'
      when engine like '%postgres%' and port = '5432' then  title || ' ' ||  engine || ' uses a default port.'
      when engine like 'oracle%' and port = '1521' then  title || ' ' ||  engine || ' uses a default port.'
      when engine like 'sqlserver%' and port = '1433' then  title || ' ' ||  engine || ' uses a default port.'
    else title || ' doesnt use a default port.'
  end as reason,
  -- Additional Dimensions                  
  region,
  account_id
from
  aws_rds_db_cluster
)
union
(
select
  -- Required Columns
  arn as resource,
  case
      when engine similar to '%(aurora|mysql|mariadb)%' and port = '3306' then 'alarm'
      when engine like '%postgres%' and port = '5432' then 'alarm'
      when engine like 'oracle%' and port = '1521' then 'alarm'
      when engine like 'sqlserver%' and port = '1433' then 'alarm'
    else 'ok'
  end as status,
  case
      when engine similar to '%(aurora|mysql|mariadb)%' and port = '3306' then  title || ' ' ||  engine || ' uses a default port.'
      when engine like '%postgres%' and port = '5432' then  title || ' ' ||  engine || ' uses a default port.'
      when engine like 'oracle%' and port = '1521' then  title || ' ' ||  engine || ' uses a default port.'
      when engine like 'sqlserver%' and port = '1433' then  title || ' ' ||  engine || ' uses a default port.'
    else title || ' doesnt use a default port.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance
);