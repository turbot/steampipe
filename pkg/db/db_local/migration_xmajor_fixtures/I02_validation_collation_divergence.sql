-- Reframed: the original fixture inserted non-ASCII
-- text, which the production pre-flight collation scan flags before the
-- post-restore validation pass gets a chance to run. The case's intent was
-- always to exercise the validation-divergence rollback path, so the fixture
-- is ASCII-only and the harness drives that path via the forceValidationFailure
-- flag on caseSetup: an invalid index is planted on the target, so the shipped
-- runValidateRestore reports an index_invalid divergence and the shared engine
-- rolls back to ValidationDiverged. That is exactly the production code path
-- I02 is meant to assert.
CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'alpha'),(2,'bravo'),(3,'charlie');
CREATE VIEW public.v AS SELECT id, name FROM public.t ORDER BY id;
