-- Reframed (exec-2c): the original fixture used non-ASCII data plus a
-- functional index over lower(name), which the production pre-flight
-- collation scan flags (lower( regex hit + non-ASCII data) before the
-- post-restore validation pass can see anything. The case's intent was always
-- to exercise the validation-rollback path when an invalid index is detected
-- post-restore. The fixture is now ASCII-only and the harness drives the
-- validation-divergence path via forceValidationFailure on caseSetup. That
-- exercises the production runValidateRestore + runCrossMajorMigration roll
-- back to ValidationDiverged, which is what I03 is meant to assert.
CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'alpha'),(2,'bravo'),(3,'charlie');
CREATE INDEX t_name_idx ON public.t (name);
