SELECT id, result, public.safe_div(20,4) AS recomputed FROM public.t ORDER BY id;
