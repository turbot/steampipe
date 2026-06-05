CREATE FUNCTION public.sumsq_sfunc(state numeric, val numeric) RETURNS numeric LANGUAGE sql IMMUTABLE AS $$ SELECT state + val*val $$;
CREATE AGGREGATE public.sumsq(numeric) (SFUNC = public.sumsq_sfunc, STYPE = numeric, INITCOND = '0');
CREATE TABLE public.t (id int, val numeric);
INSERT INTO public.t VALUES (1,2),(2,3);
