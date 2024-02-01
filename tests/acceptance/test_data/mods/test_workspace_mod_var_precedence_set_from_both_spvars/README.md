# test_workspace_mod_var_precedence_set_from_command_line_and_both_spvars

### Description

This mod is used to test variable resolution precedence in a mod by passing the --var command line arg, a steampipe.spvars file and an *.auto.spvars file. The mod also has a default value of variable 'version' set.

### Usage

This mod is used in the tests in `mod_vars.bats` to simulate a scenario where the version defined in the mod is picked from the --var command line argument over the steampipe.spvars and *.auto.spvars file and over the default value of variable 'version' set in the mod, because command line arguments have higher precendence.

Steampipe loads variables in the following order, with later sources taking precedence over earlier ones:

1. Environment variables
2. The steampipe.spvars file, if present.
3. Any *.auto.spvars files, in alphabetical order by filename.
4. Any --var and --var-file options on the command line, in the order they are provided.