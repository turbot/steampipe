select num as id, 
    case 
        when (num<=$1) then 'ok' 
        when (num>$1 and num<=$1+$2) then 'alarm'
        when (num>$1+$2 and num<=$1+$2+$3) then 'error' 
        when (num>$1+$2+$3 and num<=$1+$2+$3+$4) then 'skip' 
        when (num>$1+$2+$3+$4 and num<=$1+$2+$3+$4+$5) then 'info' 
    end status, 
    'steampipe' as resource, 
    case 
        when (num<=$1) then 'Resource satisfies condition' 
        when (num>$1 and num<=$1+$2) then 'Resource does not satisfy condition' 
        when (num>$1+$2 and num<=$1+$2+$3) then 'Resource has some error' 
        when (num>$1+$2+$3 and num<=$1+$2+$3+$4) then 'Resource is skipped' 
        when (num>$1+$2+$3+$4 and num<=$1+$2+$3+$4+$5) then 'Information' 
    end reason 
from generate_series(1, ($1::int+$2::int+$3::int+$4::int+$5::int)) num
