CREATE TABLE public.base (id int, name text);
INSERT INTO public.base VALUES (1,'café'),(2,'Zürich'),(3,'naïve');
CREATE VIEW public.v AS SELECT id, name FROM public.base ORDER BY name;
