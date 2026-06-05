-- Catalog follow-up (2026-06-05): SYSTEM_USER became reserved in PG16.
-- PG14 pg_dump emits the column name unquoted; PG18 parser rejects it (syntax error).
CREATE TABLE public.t (system_user text);
INSERT INTO public.t VALUES ('alice'),('bob');
