CREATE ROLE app_role NOLOGIN;
GRANT CREATE ON SCHEMA public TO app_role;
CREATE TABLE public.t (id int);
INSERT INTO public.t VALUES (1),(2);
