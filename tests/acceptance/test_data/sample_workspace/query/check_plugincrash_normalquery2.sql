select 
    case
        when mod(id,2)=0 then 'alarm'
        when mod(id,2)=1 then 'ok'
    end status,
    fatal_error as resource,
    retryable_error as reason
from chaos_get_errors limit 10