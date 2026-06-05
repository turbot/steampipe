CREATE FUNCTION public.add(a int, b int) RETURNS int LANGUAGE plpgsql AS $$ BEGIN RETURN a + b; END; $$;
CREATE TABLE public.t (id int, result int);
INSERT INTO public.t VALUES (1, public.add(2,3));
