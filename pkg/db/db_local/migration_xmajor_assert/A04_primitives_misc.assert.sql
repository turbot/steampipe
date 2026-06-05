SELECT id, u, encode(b,'hex') AS b, host(ip) AS ip, net::text, mac::text FROM public.t ORDER BY id;
