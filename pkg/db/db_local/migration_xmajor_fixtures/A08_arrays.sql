CREATE TABLE public.t (id int, ints int[], strs text[], grid int[][]);
INSERT INTO public.t VALUES
  (1, '{1,2,3}', '{a,b,c}', '{{1,2},{3,4}}'),
  (2, '{}', '{}', '{{9}}');
