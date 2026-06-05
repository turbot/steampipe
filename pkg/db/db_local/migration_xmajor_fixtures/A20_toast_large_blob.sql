CREATE TABLE public.t (id int, blob bytea);
INSERT INTO public.t VALUES (1, decode(repeat('ab', 4096), 'hex')), (2, decode(repeat('cd', 8192), 'hex'));
