WITH s_path AS (select setting from pg_settings where name='search_path') 
SELECT s_path.setting as resource, 
CASE 
    WHEN s_path.setting LIKE 'aws%' THEN 'ok' 
    ELSE 'alarm' 
END as status,
CASE
    WHEN s_path.setting LIKE 'aws%' THEN 'Starts with "aws"'
    ELSE 'Does not start with "aws"'
END as reason
FROM s_path