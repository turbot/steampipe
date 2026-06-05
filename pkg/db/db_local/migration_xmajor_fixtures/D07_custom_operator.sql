CREATE FUNCTION public.eq_mod10(a int, b int) RETURNS boolean LANGUAGE sql IMMUTABLE AS $$ SELECT (a % 10) = (b % 10) $$;
CREATE OPERATOR public.=== (LEFTARG = int, RIGHTARG = int, FUNCTION = public.eq_mod10);
CREATE TABLE public.t (id int, val int);
INSERT INTO public.t VALUES (1,11),(2,22);
