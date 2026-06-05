-- Catalog item P15.6: interval output stability changed (immutable -> stable) in PG15+.
-- An expression index on i::text fails to restore: "functions in index expression must be marked IMMUTABLE".
CREATE TABLE public.t (i interval);
CREATE INDEX t_i_text_idx ON public.t ((i::text));
INSERT INTO public.t VALUES (interval '1 day'),(interval '2 hours');
