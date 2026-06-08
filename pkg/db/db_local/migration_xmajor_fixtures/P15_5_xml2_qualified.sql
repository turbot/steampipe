-- Catalog item P15.5: xml_is_well_formed() moved from the xml2 contrib extension
-- to core in PG15. A view referencing the schema-qualified xml2.xml_is_well_formed()
-- form fails to restore on PG18 because the xml2 extension is not installed there
-- (CREATE EXTENSION is not a public-schema object, so --schema=public does not
-- carry it). pg_restore aborts validating the view: function does not exist.
--
-- The PG14 wrapper fixture install does not ship the xml2 extension either, so the
-- schema-qualified form cannot be created on the source. We instead reference a
-- schema-qualified function in a non-public schema that the public-schema dump will
-- NOT recreate, reproducing the identical engine path (restore aborts on a missing
-- referenced object) deterministically with only core objects.
CREATE SCHEMA helper;
CREATE FUNCTION helper.xml_is_well_formed(text) RETURNS boolean LANGUAGE sql IMMUTABLE AS $$ SELECT $1 IS NOT NULL $$;
CREATE VIEW public.v AS SELECT helper.xml_is_well_formed('<a/>') AS ok;
