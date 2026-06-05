CREATE TABLE public.t (id int, name text);
CREATE UNIQUE INDEX t_name_uidx ON public.t (name);
INSERT INTO public.t VALUES (1,'café'),(2,'Zürich');
