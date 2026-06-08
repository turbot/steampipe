-- Catalog item P16.4: the catalogue predicted that a view built via the legacy
-- CREATE RULE ... ON SELECT form fails to restore on PG16+. That predicted
-- failure is for the DIRECT-DDL-replay / pg_upgrade path, NOT the dump-and-restore
-- path Steampipe uses. By first principles of how pg_dump serialises this object:
-- attaching a "_RETURN" ON SELECT rule to a relation turns that relation into a
-- view in PostgreSQL's catalog (relkind 'v'). pg_dump inspects the catalog and
-- emits a view as an ordinary CREATE VIEW statement reconstructed from its rewrite
-- rule - it does NOT reproduce the legacy "CREATE TABLE + CREATE RULE _RETURN"
-- sequence. PG18 accepts an ordinary CREATE VIEW (views are a stable feature), so
-- on the dump-and-restore path the restore SUCCEEDS. Expected outcome:
-- AutoRestoreSucceeded - the migration engine migrates the view + its source data
-- cleanly. (Same catalogue-predicts-DDL-failure vs dump-path-succeeds shape as F04
-- GRANT RULE.)
CREATE TABLE public.src (id int, name text);
INSERT INTO public.src VALUES (1,'a'),(2,'b');
CREATE TABLE public.v (id int, name text);
CREATE RULE "_RETURN" AS ON SELECT TO public.v DO INSTEAD SELECT id, name FROM public.src;
