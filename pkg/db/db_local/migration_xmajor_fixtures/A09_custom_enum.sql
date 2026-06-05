CREATE TYPE public.mood AS ENUM ('sad','ok','happy');
CREATE TABLE public.t (id int, m public.mood);
INSERT INTO public.t VALUES (1,'happy'),(2,'sad'),(3,'ok');
