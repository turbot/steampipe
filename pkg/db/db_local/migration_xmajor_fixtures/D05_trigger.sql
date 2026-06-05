CREATE TABLE public.t (id int, name text, name_upper text);
CREATE FUNCTION public.upcase() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN NEW.name_upper := upper(NEW.name); RETURN NEW; END; $$;
CREATE TRIGGER trg BEFORE INSERT ON public.t FOR EACH ROW EXECUTE FUNCTION public.upcase();
INSERT INTO public.t (id,name) VALUES (1,'abc'),(2,'def');
