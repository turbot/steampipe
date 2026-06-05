CREATE TABLE public.t (id int, name text);
COMMENT ON TABLE public.t IS 'a documented table';
COMMENT ON COLUMN public.t.name IS 'the name column';
CREATE FUNCTION public.noop() RETURNS void LANGUAGE sql AS $$ SELECT $$;
COMMENT ON FUNCTION public.noop() IS 'does nothing';
INSERT INTO public.t VALUES (1,'a');
