-- Catalog item P15.3: pg_is_in_backup() exists on PG14, removed in PG15+.
-- PL/pgSQL bodies are NOT validated at restore time, so the function restores intact.
-- Restore SUCCEEDS; a runtime call to public.f() would error on PG18. Documents body opacity.
CREATE FUNCTION public.f() RETURNS int LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_is_in_backup(); RETURN 1; END; $$;
CREATE TABLE public.t (id int);
INSERT INTO public.t VALUES (1),(2);
