CREATE TABLE public.t (id int, doc jsonb);
CREATE INDEX t_doc_gin ON public.t USING gin (doc);
INSERT INTO public.t VALUES (1,'{"a":1}'),(2,'{"b":[1,2,3]}');
