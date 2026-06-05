CREATE TYPE public.pt AS (x int, y int);
CREATE TABLE public.t (id int, p public.pt);
INSERT INTO public.t VALUES (1, ROW(1,2)), (2, ROW(3,4));
