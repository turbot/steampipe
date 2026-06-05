CREATE TABLE public.base (id int, name text);
INSERT INTO public.base VALUES (1,'a'),(2,'b'),(3,'c');
CREATE VIEW public.v AS SELECT id, name FROM public.base WHERE id > 1;
