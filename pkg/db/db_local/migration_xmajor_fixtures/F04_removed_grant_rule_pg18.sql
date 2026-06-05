-- Catalog item P18.1: rule privileges removed from GRANT/REVOKE in PG18.
-- pg_restore aborts on a syntax error parsing GRANT RULE on PG18.
CREATE TABLE public.t (id int);
GRANT RULE ON public.t TO public;
