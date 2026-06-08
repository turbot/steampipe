-- Catalog item P16.1: pg_get_wal_records_info_till_end_of_wal() removed in PG16.
-- A user function whose PL/pgSQL body calls the removed name restores intact
-- (PG does not validate function bodies at restore time) and would only error at
-- call time. Restore SUCCEEDS; this documents body opacity (same class as F02).
CREATE FUNCTION public.f() RETURNS int LANGUAGE plpgsql AS
$$ BEGIN PERFORM pg_get_wal_records_info_till_end_of_wal('0/0'); RETURN 1; END; $$;
CREATE TABLE public.t (id int);
INSERT INTO public.t VALUES (1),(2);
