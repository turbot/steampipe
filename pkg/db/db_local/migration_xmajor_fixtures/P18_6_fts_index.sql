-- Catalog item P18.6: full-text search now uses the cluster's default collation
-- provider instead of always libc on PG18. The embedded PG14 -> PG18 clusters are
-- both initdb'd with LC_ALL=C, so a core to_tsvector GIN index restores cleanly and
-- a REINDEX is only post-upgrade hygiene, not a restore blocker. ASCII data only.
-- Restore SUCCEEDS: the case proves the engine is not tripped by the FTS provider
-- change under the C-locale cluster the migration always runs against.
CREATE TABLE public.t (id int, body text);
INSERT INTO public.t VALUES (1,'the quick brown fox'),(2,'lazy dog sleeps');
CREATE INDEX t_fts_idx ON public.t USING gin (to_tsvector('simple', body));
