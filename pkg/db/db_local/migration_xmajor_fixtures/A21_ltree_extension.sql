CREATE EXTENSION IF NOT EXISTS ltree;
CREATE TABLE public.t (id int, path public.ltree);
INSERT INTO public.t VALUES (1,'a.b.c'),(2,'a.b'),(3,'x.y.z');
