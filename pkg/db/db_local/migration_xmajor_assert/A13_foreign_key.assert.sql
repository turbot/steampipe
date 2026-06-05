SELECT c.id, c.parent_id, c.label, p.name FROM public.child c JOIN public.parent p ON c.parent_id=p.id ORDER BY c.id;
