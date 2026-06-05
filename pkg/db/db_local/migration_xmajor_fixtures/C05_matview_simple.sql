CREATE TABLE public.base (id int, val int);
INSERT INTO public.base VALUES (1,10),(2,20),(3,30);
CREATE MATERIALIZED VIEW public.mv AS SELECT id, val*2 AS doubled FROM public.base;
