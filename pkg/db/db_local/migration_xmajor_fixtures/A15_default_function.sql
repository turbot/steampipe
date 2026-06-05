CREATE TABLE public.t (
  id int,
  created timestamptz DEFAULT now(),
  u uuid DEFAULT gen_random_uuid(),
  label text DEFAULT upper('x')
);
INSERT INTO public.t (id) VALUES (1),(2),(3);
