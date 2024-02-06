select 
    case
        when mod(id,2)=0 then 'alarm'
        when mod(id,2)=1 then 'ok'
    end status,
    int8_data as resource,
    int16_data as reason
from chaos_all_numeric_column