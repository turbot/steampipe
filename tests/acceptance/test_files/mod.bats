load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

############### QUERIES ###############

@test "query with default params and no params passed through CLI" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query query.query_params_with_all_defaults --output json

  # store the reason field in `content`
  content=$(echo $output | jq '.[].reason')

  assert_equal "$content" '"default_parameter_1 default_parameter_2 default_parameter_3"'
}

@test "query with default params and some positional params passed through CLI" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query "query.query_params_with_all_defaults(\"command_param_1\")" --output json

  # store the reason field in `content`
  content=$(echo $output | jq '.[].reason')

  assert_equal "$content" '"command_param_1 default_parameter_2 default_parameter_3"'
}

@test "query with default params and some named params passed through CLI" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query "query.query_params_with_all_defaults(p1 => \"command_param_1\")" --output json

  # store the reason field in `content`
  content=$(echo $output | jq '.[].reason')

  assert_equal "$content" '"command_param_1 default_parameter_2 default_parameter_3"'
}

@test "query with no default params and no params passed through CLI" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query query.query_params_with_no_defaults --output json

  assert_output --partial 'failed to resolve args for functionality_test_mod.query.query_params_with_no_defaults: p1,p2,p3'
}

@test "query with no default params and all params passed through CLI" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query "query.query_params_with_all_defaults(\"command_param_1\",\"command_param_2\",\"command_param_3\")" --output json

  # store the reason field in `content`
  content=$(echo $output | jq '.[].reason')

  assert_equal "$content" '"command_param_1 command_param_2 command_param_3"'
}

@test "query specific array index from param - DISABLED" {
  # cd $FUNCTIONALITY_TEST_MOD
  # run steampipe query query.query_array_params_with_default --output json

  # # store the reason field in `content`
  # content=$(echo $output | jq '.[].reason')

  # assert_equal "$content" '"default_p1_element_02"'
}

@test "query with invalid param syntax" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query "query.query_map_params_with_default(\"foo \")" --output json

  # should return an error `invalid input syntax for type json`
  assert_output --partial 'invalid input syntax for type json'
  cd -
}

@test "query specific property from map param" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe query query.query_map_params_with_default --output json

  # store the reason field in `content`
  content=$(echo $output | jq '.[].reason')

  assert_equal "$content" '"default_property_value_01"'
}

############### CONTROLS ###############

@test "control with default params and no args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_defaults_and_no_args --export test.json
  echo $output
  ls

  # store the reason field in `content` 
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"default_parameter_1 default_parameter_2 default_parameter_3"'
  rm -f test.json
}

@test "control with default params and partial named args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_defaults_and_partial_named_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"default_parameter_1 command_parameter_2 default_parameter_3"'
  rm -f test.json
}

@test "control with default params and partial positional args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_defaults_and_partial_positional_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 default_parameter_2 default_parameter_3"'
  rm -f test.json
}

@test "control with default params and all named args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_defaults_and_all_named_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 command_parameter_2 command_parameter_3"'
  rm -f test.json
}

@test "control with default params and all positional args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_defaults_and_all_positional_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 command_parameter_2 command_parameter_3"'
  rm -f test.json
}

@test "control with no default params and no args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_no_defaults_and_no_args --output json

  # should return an error `failed to resolve value for 3 parameters`
  echo $output
  [ $(echo $output | grep "failed to resolve value for 3 parameters" | wc -l | tr -d ' ') -eq 0 ]
}

@test "control with no default params and all args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_with_no_defaults_with_named_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 command_parameter_2 command_parameter_3"'
  rm -f test.json
}

@test "control to access specific array index from param - DISABLED" {
  # cd $FUNCTIONALITY_TEST_MOD
  # run steampipe check control.query_params_array_with_default --export test.json

  # # store the reason field in `content`
  # content=$(cat test.json | jq '.controls[0].results[0].reason')

  # assert_equal "$content" '"default_p1_element_02"'
  # rm -f test.json
}

@test "control to access specific property from map" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_map_with_default --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"default_property_value_01"'
  rm -f test.json
}

@test "control with invaild args syntax passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_params_invalid_arg_syntax --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `invalid input syntax for type json`, so the results should be empty
  assert_equal "$content" ""
}

@test "control with inline sql with partial named args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_inline_sql_from_control_with_partial_named_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 default_parameter_2 command_parameter_3"'
  rm -f test.json
}

@test "control with inline sql with partial positional args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_inline_sql_from_control_with_partial_positional_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 command_parameter_2 default_parameter_3"'
  rm -f test.json
}

@test "control with inline sql with no args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_inline_sql_from_control_with_no_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"default_parameter_1 default_parameter_2 default_parameter_3"'
  rm -f test.json
}

@test "control with inline sql with all named args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_inline_sql_from_control_with_all_named_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 command_parameter_2 command_parameter_3"'
  rm -f test.json
}

@test "control with inline sql with all positional args passed in control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.query_inline_sql_from_control_with_all_positional_args --export test.json

  # store the reason field in `content`
  content=$(cat test.json | jq '.controls[0].results[0].reason')

  assert_equal "$content" '"command_parameter_1 command_parameter_2 command_parameter_3"'
  rm -f test.json
}

##

@test "control with neither query property nor sql property" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_no_query_no_sql --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `must define either a 'sql' property or a 'query' property`,
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control with both query property and sql property" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_both_query_and_sql --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `must define either a 'sql' property or a 'query' property`,
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control with both params property and query property" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_params_and_query --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `has 'query' property set so cannot define param blocks`,
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control referring to query with no params definitions and named args passed" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_query_with_no_def_and_named_args_passed --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error since query has o parameter definitions,
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control referring to query with no params defaults and partial positional args passed" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_insufficient_positional_args_passed --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `failed to resolve value for 3 parameters`
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control referring to query with no params defaults and partial named args passed" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_insufficient_named_args_passed --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `failed to resolve value for 3 parameters`,
  # so the results should be empty
  assert_equal "$content" ""
}

## traversal

# This test consists of a mod with nested folders, with mod.sp file within one of them(folder11).
# Running steampipe check from folder111 should give us the result since the mod.sp file is present somewhere
# up the directory tree
@test "load a mod from an arbitrarily nested sub folder - PASS" {
  # go to the nested sub directory within the mod
  cd $FILE_PATH/test_data/nested_mod/folder1/folder11/folder111

  run steampipe check all
  assert_success
  cd -
}

# This test consists of a mod with nested folders, with mod.sp file within one of them(folder11).
# Running steampipe check from folder1(i.e. _above_ the mod folder) should return an error, since the mod.sp file is present nowhere
# up the directory tree
@test "load a mod from an arbitrarily nested sub folder - FAIL" {
  # go to the nested sub directory within the mod
  cd $FILE_PATH/test_data/nested_mod/folder1

  run steampipe check all
  assert_equal "$output" "Error: This command requires a mod definition file(mod.sp) - could not find in the current directory tree.

You can either clone a mod repository or install a mod using steampipe mod install and run this command from the cloned/installed mod directory.
Please refer to: https://steampipe.io/docs/mods/overview"
  cd -
}

# This test consists of a mod with nested folders, with no mod.sp file in any of them.
# Running steampipe check from folder11 should return an error, since the mod.sp file is present nowhere
# up the directory tree
# Running steampipe query from folder11 should give us the result since query is independent of mod.sp file.
@test "check and query from an arbitrarily nested sub folder - PASS & FAIL" {
  # go to the nested sub directory within the mod
  cd $FILE_PATH/test_data/nested_mod_no_mod_file/folder1/folder11

  run steampipe check all
  assert_equal "$output" "Error: This command requires a mod definition file(mod.sp) - could not find in the current directory tree.

You can either clone a mod repository or install a mod using steampipe mod install and run this command from the cloned/installed mod directory.
Please refer to: https://steampipe.io/docs/mods/overview"

  run steampipe query control.check_1
  assert_success
  cd -
}

## parsing

@test "mod parsing" {
  # install necessary plugins
  steampipe plugin install aws oci azure azuread

  # create a directory to install the mods
  target_directory=$(mktemp -d)
  cd $target_directory

  # install steampipe-mod-aws-compliance
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-aws-thrifty
  steampipe mod install github.com/turbot/steampipe-mod-aws-thrifty
  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-aws-thrifty@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-oci-compliance
  steampipe mod install github.com/turbot/steampipe-mod-oci-compliance
  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-oci-compliance@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-azure-compliance
  steampipe mod install github.com/turbot/steampipe-mod-azure-compliance
  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-azure-compliance@*
  run steampipe query "select 1"
  assert_success
  cd -

  # remove the directory
  cd ..
  rm -rf $target_directory
  
  
  # remove the connection config files
  rm -f $STEAMPIPE_INSTALL_DIR/config/aws.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/ibm.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/oci.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/azure.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/azuread.spc
  
  # uninstall the plugins
  steampipe plugin uninstall aws oci azure azuread

  # rerun steampipe to make sure they are removed from steampipe
  steampipe query "select 1"
}

## dependency resolution tests

@test "complex mod dependency resolution - test vars resolution from require section of local mod" {
  cd $FILE_PATH/test_data/local_mod_with_args_in_require
  steampipe mod install

  run steampipe query dependency_vars_1.query.version --output csv
  # check the output - query should use the value of variable from the local 
  # mod require section("v3.0.0") which will give the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | v3.0.0 | v3.0.0   | ok     |
# +--------+----------+--------+
  assert_output 'reason,resource,status
v3.0.0,v3.0.0,ok'
  
  rm -rf .steampipe/
  rm -rf .mod.cache.json
}
