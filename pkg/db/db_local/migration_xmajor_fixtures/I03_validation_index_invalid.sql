-- Reframed: the original fixture used non-ASCII
-- data plus a functional index over lower(name), which the production
-- pre-flight collation scan flags (lower( regex hit + non-ASCII data) before
-- the post-restore validation pass can see anything. The case's intent was
-- always to exercise the validation-rollback path when an invalid index is
-- detected post-restore - which is now driven literally: the harness's
-- forceValidationFailure plants a genuinely INVALID index on the target, the
-- shipped runValidateRestore reports index_invalid, and the shared engine
-- rolls back to ValidationDiverged. That is what I03 is meant to assert.
CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'alpha'),(2,'bravo'),(3,'charlie');
CREATE INDEX t_name_idx ON public.t (name);
