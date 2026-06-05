CREATE TABLE public.t (id int, amount numeric, ts timestamptz, doc jsonb);
INSERT INTO public.t SELECT g, (g % 1000)::numeric / 7, now() - (g || ' seconds')::interval, jsonb_build_object('n', g)
  FROM generate_series(1,1000000) g;
