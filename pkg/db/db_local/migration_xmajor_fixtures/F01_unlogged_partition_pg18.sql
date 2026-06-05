-- Catalog item P18.3: unlogged partitioned tables disallowed in PG18.
-- PG14 accepts CREATE UNLOGGED TABLE ... PARTITION BY; pg_restore aborts on PG18:
-- "partitioned tables cannot be unlogged".
-- (Substitutes the original P15.4 plpython2u case: plpython2u is not available in the
-- wrapper-built PG14 fixture install, so the fixture could not be applied on the source.)
CREATE UNLOGGED TABLE public.t (id int, d date) PARTITION BY RANGE (d);
CREATE UNLOGGED TABLE public.t_2020 PARTITION OF public.t FOR VALUES FROM ('2020-01-01') TO ('2021-01-01');
INSERT INTO public.t VALUES (1,'2020-06-01'),(2,'2020-09-15');
