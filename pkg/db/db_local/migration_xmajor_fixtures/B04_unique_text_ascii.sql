CREATE TABLE public.t (id int, code text);
CREATE UNIQUE INDEX t_code_uidx ON public.t (code);
INSERT INTO public.t VALUES (1,'AAA'),(2,'BBB'),(3,'CCC');
