CREATE TABLE public.t (id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY, name text);
INSERT INTO public.t (name) VALUES ('a'),('b'),('c');
