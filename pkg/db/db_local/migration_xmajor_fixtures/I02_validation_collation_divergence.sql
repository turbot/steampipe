-- Reframed (exec-2c): the original fixture inserted non-ASCII text into a
-- table, which the production pre-flight collation scan flags before the
-- post-restore validation pass gets a chance to run. The case's intent was
-- always to exercise the validation-divergence rollback path, so the fixture
-- is now ASCII-only and the harness drives that path via the
-- forceValidationFailure flag on caseSetup (the new conn is pointed at an
-- empty database, so runValidateRestore reports row-count divergence on every
-- table). That is exactly the production code path I02 was meant to assert.
CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'alpha'),(2,'bravo'),(3,'charlie');
CREATE VIEW public.v AS SELECT id, name FROM public.t ORDER BY id;
