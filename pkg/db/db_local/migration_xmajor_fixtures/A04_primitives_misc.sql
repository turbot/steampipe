CREATE TABLE public.t (id int, u uuid, b bytea, ip inet, net cidr, mac macaddr);
INSERT INTO public.t VALUES
  (1, '11111111-1111-1111-1111-111111111111', '\xDEADBEEF', '192.168.0.1', '10.0.0.0/8', '08:00:2b:01:02:03'),
  (2, '22222222-2222-2222-2222-222222222222', '\x00', '::1', '2001:db8::/32', '08:00:2b:01:02:04');
