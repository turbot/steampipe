CREATE TABLE public.a (id int, label text);
CREATE TABLE public.b (a_id int, qty int);
INSERT INTO public.a VALUES (1,'x'),(2,'y');
INSERT INTO public.b VALUES (1,10),(1,5),(2,7);
CREATE VIEW public.v AS SELECT a.label, sum(b.qty) AS total FROM public.a JOIN public.b ON a.id=b.a_id GROUP BY a.label;
