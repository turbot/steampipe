-- Catalog item P18.4: COPY FROM no longer treats \. as EOF in CSV on PG18. This
-- only affects CSV/text-format COPY data streams. The migration dumps in custom
-- (public) / directory (data-tank) format, where row data is not parsed as CSV, so
-- a literal "\." inside a text value round-trips cleanly. The case proves the
-- engine's binary/custom data path is not tripped by the CSV EOF change: Restore
-- SUCCEEDS, and the row containing "\." survives byte-for-byte.
CREATE TABLE public.t (id int, payload text);
INSERT INTO public.t VALUES (1, 'before'), (2, E'a\\.b'), (3, 'after');
