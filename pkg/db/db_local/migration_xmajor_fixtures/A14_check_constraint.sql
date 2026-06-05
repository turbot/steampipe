CREATE TABLE public.t (
  id int CHECK (id > 0),
  amount numeric CHECK (amount >= 0),
  name text CHECK (length(name) > 0),
  created timestamp CHECK (created > '2000-01-01')
);
INSERT INTO public.t VALUES (1, 100, 'ok', '2021-01-01');
