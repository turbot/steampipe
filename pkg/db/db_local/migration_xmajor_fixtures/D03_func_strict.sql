CREATE FUNCTION public.safe_div(a int, b int) RETURNS int LANGUAGE sql STRICT IMMUTABLE AS $$ SELECT a / b $$;
CREATE TABLE public.t (id int, result int);
INSERT INTO public.t VALUES (1, public.safe_div(10,2));
