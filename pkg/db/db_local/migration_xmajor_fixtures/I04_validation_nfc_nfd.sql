CREATE TABLE public.t (id int, name text);
-- 'é' as precomposed (NFC, U+00E9) vs decomposed (NFD, e + U+0301)
INSERT INTO public.t VALUES (1, E'caf\u00e9'), (2, E'cafe\u0301'), (3, 'cafz');
CREATE VIEW public.v AS SELECT id, name FROM public.t ORDER BY name;
