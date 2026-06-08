-- Catalog item P16.2: pg_get_wal_stats_till_end_of_wal() removed in PG16.
-- Same shape as P16.1 for the stats variant: the function body referencing the
-- removed name restores intact; a runtime call would error on PG18. Restore
-- SUCCEEDS (body opacity).
CREATE FUNCTION public.f() RETURNS int LANGUAGE plpgsql AS
$$ BEGIN PERFORM pg_get_wal_stats_till_end_of_wal('0/0', '0/0', true); RETURN 1; END; $$;
CREATE TABLE public.t (id int);
INSERT INTO public.t VALUES (3),(4);
