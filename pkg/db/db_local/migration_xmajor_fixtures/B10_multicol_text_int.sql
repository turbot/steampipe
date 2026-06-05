CREATE TABLE public.t (id int, name text, rank int);
CREATE INDEX t_multi_idx ON public.t (name, rank);
INSERT INTO public.t VALUES (1,'café',10),(2,'Zürich',20),(3,'naïve',30);
