CREATE TABLE public.t (id int, n bigint);
CREATE INDEX t_n_idx ON public.t (n);
CREATE UNIQUE INDEX t_id_uidx ON public.t (id);
INSERT INTO public.t VALUES (1,100),(2,200);
