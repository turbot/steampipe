-- H04: deliberately-incompatible object forces a mid-transaction restore abort.
-- Uses the verified P18.3 failure (unlogged partitioned table) so pg_restore
-- --single-transaction aborts and the non-fatal degrade path is exercised.
CREATE TABLE public.keep (id int, name text);
INSERT INTO public.keep VALUES (1,'a'),(2,'b');
CREATE UNLOGGED TABLE public.part (id int, d date) PARTITION BY RANGE (d);
CREATE UNLOGGED TABLE public.part_2020 PARTITION OF public.part FOR VALUES FROM ('2020-01-01') TO ('2021-01-01');
INSERT INTO public.part VALUES (1,'2020-06-01');
