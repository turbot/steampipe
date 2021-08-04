WITH s_path AS (select setting from pg_settings where name='search_path') 
SELECT s_path.setting as resource, 
CASE 
    WHEN s_path.setting LIKE 'a, b, c%' THEN 'ok' 
    ELSE 'alarm' 
END as status, '' as reason
FROM s_path