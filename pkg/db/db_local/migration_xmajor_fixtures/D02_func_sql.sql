CREATE FUNCTION public.dbl(a int) RETURNS int LANGUAGE sql AS $$ SELECT a * 2 $$;
CREATE TABLE public.t (id int, result int);
INSERT INTO public.t VALUES (1, public.dbl(21));
