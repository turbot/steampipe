CREATE TABLE public.t (id int, name text, active boolean);
CREATE INDEX t_name_partial ON public.t (name) WHERE active;
INSERT INTO public.t VALUES (1,'café',true),(2,'Zürich',true),(3,'naïve',false);
