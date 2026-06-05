CREATE TABLE public.t (id int, region text) PARTITION BY LIST (region);
CREATE TABLE public.t_us PARTITION OF public.t FOR VALUES IN ('us');
CREATE TABLE public.t_eu PARTITION OF public.t FOR VALUES IN ('eu');
INSERT INTO public.t VALUES (1,'us'),(2,'eu'),(3,'us');
