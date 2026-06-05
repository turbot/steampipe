CREATE TABLE public.t (id int, ts timestamp, tstz timestamptz, d date, tm time, iv interval);
INSERT INTO public.t VALUES
  (1, '2021-01-02 03:04:05', '2021-01-02 03:04:05+00', '2021-01-02', '03:04:05', interval '1 day 2 hours'),
  (2, '1999-12-31 23:59:59', '2000-01-01 00:00:00+05', '1999-12-31', '23:59:59', interval '90 minutes');
