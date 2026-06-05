SELECT count(*) AS rows, min(id) AS lo, max(id) AS hi, md5(string_agg(name, ',' ORDER BY id)) AS digest FROM public.t;
