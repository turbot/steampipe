CREATE TABLE public.t (
  id int,
  w int,
  h int,
  area int GENERATED ALWAYS AS (w * h) STORED
);
INSERT INTO public.t (id,w,h) VALUES (1,2,3),(2,4,5);
