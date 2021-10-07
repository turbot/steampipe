query "chaos6_query"{
    title ="chaos6_query"
    description = "Query using the chaos6 connection which contains the options block to verify parsing"
    sql = "select case when mod(id,2)=0 then 'alarm' when mod(id,2)=1 then 'ok' end status, time_now as resource, id as reason from chaos6.chaos_cache_check where id=2"
}