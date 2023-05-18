package db_common

// Functions is a list of SQLFunction objects that are installed in the db 'steampipe_internal' schema startup
var Functions = []SQLFunction{
	{
		Name:     "glob",
		Params:   map[string]string{"input_glob": "text"},
		Returns:  "text",
		Language: "plpgsql",
		Body: `
declare
	output_pattern text;
begin
	output_pattern = replace(input_glob, '*', '%');
	output_pattern = replace(output_pattern, '?', '_');
	return output_pattern;
end;
`,
	},
	{
		Name:     "set_cache",
		Params:   map[string]string{"command": "text"},
		Returns:  "void",
		Language: "plpgsql",
		Body: `
begin
	IF command = 'on' THEN
		INSERT INTO steampipe_internal.steampipe_settings("name","value") VALUES ('cache','true');
	ELSIF command = 'off' THEN
		INSERT INTO steampipe_internal.steampipe_settings("name","value") VALUES ('cache','false');
	ELSIF command = 'clear' THEN
		INSERT INTO steampipe_internal.steampipe_settings("name","value") VALUES ('cache_clear_time','');
	ELSE
		RAISE EXCEPTION 'Unknown value % for set_cache - valid values are on, off and clear.', $1;
	END IF;
end;
`,
	},
}
