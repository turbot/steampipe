CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'café'),(2,'cafe'),(3,'Café');
CREATE INDEX t_name_lower_idx ON public.t (lower(name) text_pattern_ops);
