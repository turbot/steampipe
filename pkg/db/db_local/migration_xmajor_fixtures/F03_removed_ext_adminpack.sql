-- Catalog item P15.3 (substitute for P17.1 adminpack, which the PG14 fixture install lacks):
-- pg_is_in_backup() exists on PG14 but was removed in PG15+. A VIEW referencing it is validated
-- at restore time, so pg_restore aborts on PG18: "function pg_is_in_backup() does not exist".
CREATE VIEW public.v AS SELECT pg_is_in_backup() AS in_backup;
