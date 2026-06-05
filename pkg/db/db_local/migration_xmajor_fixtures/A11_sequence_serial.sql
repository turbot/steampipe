CREATE TABLE public.t (id serial PRIMARY KEY, big bigserial, name text);
INSERT INTO public.t (name) VALUES ('a'),('b'),('c');
