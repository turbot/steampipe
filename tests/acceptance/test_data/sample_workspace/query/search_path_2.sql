WITH s_path AS (select setting from pg_settings where name='search_path') 
SELECT s_path.setting as resource, 
CASE 
    WHEN s_path.setting LIKE 'a, b, c%' THEN 'ok' 
    ELSE 'alarm' 
END as status,
CASE
    WHEN s_path.setting LIKE 'aws%' THEN 'Starts with "a, b, c"'
    ELSE 'Does not start with "a, b, c"'
END as reason
FROM s_path