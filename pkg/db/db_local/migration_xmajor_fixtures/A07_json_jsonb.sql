CREATE TABLE public.t (id int, j json, jb jsonb);
INSERT INTO public.t VALUES
  (1, '{"a":1,"b":[2,3]}', '{"nested":{"x":true},"arr":[1,2,3]}'),
  (2, '[]', '{}');
