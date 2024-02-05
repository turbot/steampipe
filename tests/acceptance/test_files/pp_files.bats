load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

### ppvars file tests ###

@test "test variable resolution in workspace mod set from *.auto.ppvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_auto.ppvars

  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the *.auto.ppvars
  # file ("v7.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v7.0.0 | v7.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v7.0.0,v7.0.0,ok'
}

@test "test variable resolution in workspace mod set from explicit ppvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_explicit_ppvars

  run steampipe query query.version --output csv --var-file='deps.ppvars'
  # check the output - query should use the value of variable set from the explicit ppvars
  # file ("v8.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v8.0.0 | v8.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v8.0.0,v8.0.0,ok'
}

@test "test variable resolution in dependency mod set from *.auto.ppvars file" {
  cd $FILE_PATH/test_data/mods/test_dependency_mod_var_set_from_auto.ppvars

  run steampipe query dependency_vars_1.query.version --output csv
  # check the output - query should use the value of variable set from the *.auto.ppvars 
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

@test "test variable resolution precedence in workspace mod set from auto.ppvars and ENV" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_set_from_auto.ppvars
  export SP_VAR_version=v9.0.0
  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the *.auto.ppvars("v7.0.0") file over 
  # ENV("v9.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v7.0.0 | v7.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v7.0.0,v7.0.0,ok'
}

@test "test variable resolution precedence in workspace mod set from command line(--var) and steampipe.ppvars file and *.auto.ppvars file" {
  cd $FILE_PATH/test_data/mods/test_workspace_mod_var_precedence_set_from_both_ppvars

  run steampipe query query.version --output csv --var version="v5.0.0"
  # check the output - query should use the value of variable set from the command line --var flag("v5.0.0") over 
  # steampipe.ppvars("v7.0.0") and *.auto.ppvars file("v8.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v5.0.0 | v5.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v5.0.0,v5.0.0,ok'
}

### mod.pp file tests ###

@test "test that mod.pp is not renamed after uninstalling mod" {
  cd $FILE_PATH/test_data/mods/local_mod_with_mod.pp_file

  run steampipe mod install
  assert_success

  run steampipe mod uninstall
  # check mod.pp file still exists and is not renamed
  run ls mod.pp
  assert_success
}
