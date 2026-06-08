-- Catalog item P17.5: pg_attribute.attstattarget semantic shift in PG17 (NULL now
-- means "default"; was -1). A view filtering WHERE attstattarget = -1 RESTORES
-- successfully -- the column still exists -- but returns different rows on PG18.
-- This is a behavioural divergence, NOT a restore failure, so the migration engine
-- must let it through: Restore SUCCEEDS. No row-comparing assert golden is attached
-- because the divergence is exactly that PG14 and PG18 return different rows; the
-- case asserts only that the engine is not tripped into a failure outcome.
CREATE VIEW public.v AS
  SELECT attrelid, attname FROM pg_attribute WHERE attstattarget = -1;
