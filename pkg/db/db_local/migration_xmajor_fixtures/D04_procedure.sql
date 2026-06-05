CREATE TABLE public.t (id int, name text);
CREATE PROCEDURE public.seed() LANGUAGE plpgsql AS $$ BEGIN INSERT INTO public.t VALUES (1,'seeded'); END; $$;
CALL public.seed();
