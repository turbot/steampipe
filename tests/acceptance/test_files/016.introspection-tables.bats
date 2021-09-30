load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "ensure mod name in introspection table is <mod_name> not mod.<mod_name>" {
  cd $WORKSPACE_DIR
  run steampipe query "select * from steampipe_query" --output json
  
  # extract the first mod_name from the list
  mod_name=$(echo $output | jq '.[0].mod_name')

  # check if mod_name starts with "mod."
  if [[ "$mod_name" == *"mod."* ]];
  then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "0"
}

@test "ensure query pseudo resources, i.e. sql files, have resource name <query_name> not <query.query_name>" {
  cd $WORKSPACE_DIR
  run steampipe query "select * from steampipe_query" --output json

  # extract the first encountered sql file's file_name from the list
  sql_file_name=$(echo $output | jq '.[].file_name' | grep ".sql" | head -1)

  #extract the resource_name of the above extracted file_name
  resource_name=$(echo $output | jq --arg FILENAME "$sql_file_name" '.[] | select(.file_name=="$FILENAME") | .resource_name')

  # check if resource_name starts with "query."
  if [[ "$resource_name" == *"query."* ]];
  then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "0"
}

@test "ensure the referenced_by column is populated correctly" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_query" --output json

  # extract the refs and the referenced_by
  refs=$(echo $output | jq '.[].refs' | tr -d '[:space:]')
  referenced_by=$(echo $output | jq '.[].referenced_by' | tr -d '[:space:]')

  assert_equal "$refs" '["var.sample_var_1"]'
  assert_equal "$referenced_by" '["control.sample_control_1"]'
}

@test "ensure the refs column includes variable references" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_query" --output json

  # extract the refs
  refs=$(echo $output | jq '.[].refs' | tr -d '[:space:]')
  echo $refs

  # check if refs contains variables(start with "var.")
  if [[ "$refs" == *"var."* ]];
  then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "1"
}

@test "introspection tables should get populated in query batch mode" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_query" --output json
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_table.json)"
}