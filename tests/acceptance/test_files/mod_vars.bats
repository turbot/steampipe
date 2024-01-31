load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

### workspace mod tests ###

@test "test variable resolution in workspace mod set from command line(--var)" {
  cd $FILE_PATH/test_data/mods/test_vars_workspace_mod

  run steampipe query query.version --output csv --var version="v5.0.0"
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

@test "test variable resolution in workspace mod set from auto spvars file" {
  cd $FILE_PATH/test_data/mods/test_vars_workspace_mod

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
  cd $FILE_PATH/test_data/mods/test_vars_workspace_mod

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
  cd $FILE_PATH/test_data/mods/test_vars_workspace_mod
  export SP_VAR_version=v9.0.0
  run steampipe query query.version --output csv
  # check the output - query should use the value of variable set from the ENV var
  # SP_VAR_top ("v9.0.0") which will give the output:
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
  cd $FILE_PATH/test_data/mods/test_vars_dependency_mod

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

@test "test variable resolution in dependency mod set from auto spvars file" {
  cd $FILE_PATH/test_data/mods/test_vars_dependency_mod

  run steampipe query dependency_vars_1.query.version --output csv
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

@test "test variable resolution in dependency mod set from explicit spvars file" {
  cd $FILE_PATH/test_data/mods/test_vars_dependency_mod

  run steampipe query dependency_vars_1.query.version --output csv --var-file='deps.spvars'
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

@test "test variable resolution in dependency mod set from ENV" {
  cd $FILE_PATH/test_data/mods/test_vars_dependency_mod
  export SP_VAR_version=v9.0.0
  run steampipe query dependency_vars_1.query.version --output csv
  # check the output - query should use the value of variable set from the ENV var
  # SP_VAR_top ("v9.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v9.0.0 | v9.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v9.0.0,v9.0.0,ok'
}
