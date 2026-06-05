CREATE TABLE public.t (id int, name text);
CREATE INDEX t_lower_idx ON public.t (lower(name));
INSERT INTO public.t VALUES (1,'café'),(2,'Zürich'),(3,'naïve');
