# test_workspace_mod_var_set_from_command_line

### Description

This mod is used to test variable resolution in a mod by passing the --var command line arg. The mod has a default value of variable 'version' set.

### Usage

This mod is used in the tests in `mod_vars.bats` to simulate a scenario where the version defined in the mod is picked from the passed
command line argument over the default value of variable 'version' set in the mod.