CREATE TABLE public.parent (id int PRIMARY KEY, name text);
CREATE TABLE public.child (id int PRIMARY KEY, parent_id int REFERENCES public.parent(id), label text);
INSERT INTO public.parent VALUES (1,'p1'),(2,'p2');
INSERT INTO public.child VALUES (10,1,'c1'),(20,1,'c2'),(30,2,'c3');
