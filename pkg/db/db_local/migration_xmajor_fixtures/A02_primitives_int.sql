CREATE TABLE public.t (id int PRIMARY KEY, big bigint, num numeric(12,4), flag boolean);
INSERT INTO public.t VALUES (1, 9223372036854775807, 3.1415, true), (2, -10, -0.0001, false), (3, 0, 0, NULL);
