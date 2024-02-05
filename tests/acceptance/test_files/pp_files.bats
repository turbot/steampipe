load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"


@test "test variable resolution in workspace mod set from auto spvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_auto_spvars

  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the auto spvars
  # file ("v7.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v7.0.0 | v7.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v7.0.0,v7.0.0,ok'
}

@test "test variable resolution in workspace mod set from explicit spvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_explicit_spvars

  run steampipe query query.version --output csv --var-file='deps.spvars'
  # check the output - query should use the value of variable set from the explicit spvars
  # file ("v8.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v8.0.0 | v8.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v8.0.0,v8.0.0,ok'
}

@test "test variable resolution in workspace mod set from ENV" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_command_line
  export SP_VAR_version=v9.0.0
  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the ENV var
  # SP_VAR_version ("v9.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v9.0.0 | v9.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v9.0.0,v9.0.0,ok'
}

### dependency mod tests ###
# The following set of tests use a dependency mod(the mod is committed) that has a variable dependency but the
# variable does not have a default. This means that the variable must be set from the command
# line, an auto spvars file, an explicit spvars file, or an ENV var. The tests below check that
# the variable is resolved correctly in each of these cases.

@test "test variable resolution in dependency mod set from command line(--var)" {
  cd $FILE_PATH/test_data/mods/test_dependency_mod_var_set_from_command_line

  run steampipe query dependency_vars_1.query.version --output csv --var dependency_vars_1.version="v5.0.0"
  # check the output - query should use the value of variable set from the command line
  # --var flag ("v5.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v5.0.0 | v5.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v5.0.0,v5.0.0,ok'
}

@test "test variable resolution in dependency mod set from steampipe.spvars file" {
  cd $FILE_PATH/test_data/mods/test_dependency_mod_var_set_from_auto_spvars

  run steampipe query dependency_vars_1.query.version --output csv
  # check the output - query should use the value of variable set from the steampipe.spvars
  # file ("v7.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v7.0.0 | v7.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v7.0.0,v7.0.0,ok'
}

@test "test variable resolution in dependency mod set from *.auto.spvars spvars file" {
  cd $FILE_PATH/test_data/mods/test_dependency_mod_var_set_from_explicit_spvars

  run steampipe query dependency_vars_1.query.version --output csv
  # check the output - query should use the value of variable set from the *.auto.spvars 
  # file ("v8.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v8.0.0 | v8.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v8.0.0,v8.0.0,ok'
}

### precedence tests ###

@test "test variable resolution precedence in workspace mod set from steampipe.spvars and *.auto.spvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_precedence_set_from_both_spvars

  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the  *.auto.spvars("v8.0.0") file over 
  # steampipe.spvars("v7.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v8.0.0 | v8.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v8.0.0,v8.0.0,ok'
}

@test "test variable resolution precedence in workspace mod set from steampipe.spvars and ENV" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_auto_spvars
  export SP_VAR_version=v9.0.0
  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the steampipe.spvars("v7.0.0") file over 
  # ENV("v9.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v7.0.0 | v7.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v7.0.0,v7.0.0,ok'
}

@test "test variable resolution precedence in workspace mod set from command line(--var) and steampipe.spvars file and *.auto.spvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_precedence_set_from_both_spvars

  run steampipe query query.version --output csv --var version="v5.0.0"
  # check the output - query should use the value of variable set from the command line --var flag("v5.0.0") over 
  # steampipe.spvars("v7.0.0") and *.auto.spvars file("v8.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v5.0.0 | v5.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v5.0.0,v5.0.0,ok'
}

