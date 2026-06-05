CREATE TABLE public.t (id int, val int, d date);
CREATE INDEX t_val_idx ON public.t (val);
INSERT INTO public.t VALUES (1,10,'2021-01-01');
