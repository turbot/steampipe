-- Catalog item P18.2: pg_backend_memory_contexts.parent column removed in PG18.
-- A user view selecting parent from this catalog view fails to restore with
-- "column parent does not exist". Expected outcome: RestoreFailedGracefully
-- (dump + old dir retained).
CREATE VIEW public.v AS SELECT name, parent FROM pg_backend_memory_contexts;
