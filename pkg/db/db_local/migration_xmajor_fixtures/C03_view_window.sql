CREATE TABLE public.base (id int, grp int, val int);
INSERT INTO public.base VALUES (1,1,10),(2,1,20),(3,2,30);
CREATE VIEW public.v AS SELECT id, grp, val, row_number() OVER (PARTITION BY grp ORDER BY val) AS rn FROM public.base;
