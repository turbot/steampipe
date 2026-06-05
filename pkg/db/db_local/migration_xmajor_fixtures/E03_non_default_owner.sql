CREATE ROLE owner_role NOLOGIN;
CREATE TABLE public.t (id int, name text);
ALTER TABLE public.t OWNER TO owner_role;
INSERT INTO public.t VALUES (1,'a'),(2,'b');
