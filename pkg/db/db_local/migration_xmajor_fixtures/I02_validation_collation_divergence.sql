CREATE TABLE public.t (id int, name text);
INSERT INTO public.t VALUES (1,'cote'),(2,E'cot\u00e9'),(3,E'c\u00f4te'),(4,E'c\u00f4t\u00e9');
-- ORDER BY id (integer) so the pre-flight text-order-view heuristic does not flag this case;
-- the collation-sensitive content is in name, surfaced only by the validation pass.
CREATE VIEW public.v AS SELECT id, name FROM public.t ORDER BY id;
