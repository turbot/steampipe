SELECT id, result, public.add(10,20) AS recomputed FROM public.t ORDER BY id;
