select 
    case
        when num=1 then 'ok'
        when mod(num,2)=0 then 'alarm'
        when mod(num,3)=0 then 'ok'
        when mod(num,5)=0 then 'error'
        when mod(num,7)=0 then 'info'
        when mod(num,11)=0 then 'skip'
    end status,
    'steampipe' as resource,
    case
        when num=1 then 'ok'
        when mod(num,2)=0 then 'alarm'
        when mod(num,3)=0 then 'ok'
        when mod(num,5)=0 then 'error'
        when mod(num,7)=0 then 'info'
        when mod(num,11)=0 then 'skip'
    end reason
from generate_series(1, 12) num