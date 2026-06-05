CREATE TABLE public.safe (id int, n int);
CREATE INDEX safe_n_idx ON public.safe (n);
CREATE TABLE public.risky (id int, name text);
CREATE INDEX risky_name_idx ON public.risky (name);
INSERT INTO public.safe VALUES (1,10);
INSERT INTO public.risky VALUES (1,'café');
