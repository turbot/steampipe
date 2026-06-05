CREATE TABLE public.t (id int, name text);
CREATE INDEX t_name_idx ON public.t (name);
INSERT INTO public.t VALUES (1,'café'),(2,'日本語');
