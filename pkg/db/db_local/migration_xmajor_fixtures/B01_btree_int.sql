CREATE TABLE public.t (id int, val int);
CREATE INDEX t_val_idx ON public.t (val);
INSERT INTO public.t SELECT g, g*2 FROM generate_series(1,100) g;
