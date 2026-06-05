SELECT id, length(blob) AS len, md5(blob) AS digest FROM public.t ORDER BY id;
