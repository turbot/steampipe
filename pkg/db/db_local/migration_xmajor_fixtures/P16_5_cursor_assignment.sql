-- Catalog item P16.5: PL/pgSQL bound-cursor assignment semantics changed in PG16
-- (the string is bound at OPEN time, not at the assignment point). The function
-- body restores intact; only runtime behaviour differs. Restore SUCCEEDS (body
-- opacity) - the migration engine must not be tripped by a behavioural-only change.
CREATE FUNCTION public.f() RETURNS refcursor LANGUAGE plpgsql AS $$
DECLARE c refcursor;
BEGIN
  c := 'mycur';
  RETURN c;
END; $$;
CREATE TABLE public.t (id int);
INSERT INTO public.t VALUES (5),(6);
