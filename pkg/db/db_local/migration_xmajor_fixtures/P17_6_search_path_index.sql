-- Catalog item P17.6: functions invoked from index expressions run with a safe
-- search_path during maintenance on PG17+. A PG14 expression index built on a
-- user function restores successfully; only a later REINDEX may resolve the
-- function differently. This is a maintenance-time behavioural divergence, NOT a
-- restore failure. ASCII data only, so the collation pre-flight scan does not fire
-- and the case isolates the search_path risk: Restore SUCCEEDS.
CREATE FUNCTION public.myupper(t text) RETURNS text LANGUAGE sql IMMUTABLE AS $$ SELECT upper(t) $$;
CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'alice'),(2,'bob');
CREATE INDEX t_name_idx ON public.t (public.myupper(name));
