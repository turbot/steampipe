CREATE TABLE public.t (id int, name text);
INSERT INTO public.t SELECT g, 'row-' || g FROM generate_series(1,100000) g;
