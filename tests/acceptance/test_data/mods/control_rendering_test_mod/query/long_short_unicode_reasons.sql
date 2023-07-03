select 
    case
        when mod(num,2)=0 then 'alarm'
        when mod(num,3)=0 then 'ok'
        when mod(num,5)=0 then 'error'
    end status,
    'steampipe' as resource,
    case
        when mod(num,2)=0 then 'alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm alarm'
        when mod(num,3)=0 then 'ok'
        when mod(num,5)=0 then 'error ❌'
    end reason
from generate_series(2, 5) num