CREATE TABLE public.t (id int, name text);
CREATE INDEX t_name_idx ON public.t (name);
INSERT INTO public.t VALUES (1,'alpha'),(2,'bravo'),(3,'charlie');
