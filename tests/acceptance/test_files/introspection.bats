load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe_introspection=none" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=none
  run steampipe query "select * from steampipe_query" --output json

  assert_output --partial 'relation "steampipe_query" does not exist'
}

@test "resource=query | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_query" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 8th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "8d" output.json
  else
    run sed -i "8d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_query.json)"
  rm -f output.json*
}

@test "resource=control | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_control" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 11th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "11d" output.json
  else
    run sed -i "11d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_control.json)"
  rm -f output.json*
}

@test "resource=variable | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_variable" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 8th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "8d" output.json
    run sed -i ".json" "33d" output.json
  else
    run sed -i "8d" output.json
    run sed -i "33d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_variable.json)"
  rm -f output.json*
}

@test "resource=benchmark | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_benchmark" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 10th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "10d" output.json
  else
    run sed -i "10d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_benchmark.json)"
  rm -f output.json*
}

@test "resource=dashboard | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 11th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "11d" output.json
  else
    run sed -i "11d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard.json)"
  rm -f output.json*
}

@test "resource=dashboard_card | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_card" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 8th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "8d" output.json
  else
    run sed -i "8d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_card.json)"
  rm -f output.json*
}

@test "resource=dashboard_image | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_image" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 9th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "9d" output.json
  else
    run sed -i "9d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_image.json)"
  rm -f output.json*
}

@test "resource=dashboard_text | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_text" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 7th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "7d" output.json
  else
    run sed -i "7d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_text.json)"
  rm -f output.json*
}

@test "resource=dashboard_chart | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_chart" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 9th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "9d" output.json
  else
    run sed -i "9d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_chart.json)"
  rm -f output.json*
}

@test "resource=dashboard_flow | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_flow" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 13th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "13d" output.json
  else
    run sed -i "13d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_flow.json)"
  rm -f output.json*
}

@test "resource=dashboard_graph | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_graph" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 14th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "14d" output.json
  else
    run sed -i "14d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_graph.json)"
  rm -f output.json*
}

@test "resource=dashboard_hierarchy | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_hierarchy" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 13th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "13d" output.json
  else
    run sed -i "13d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_hierarchy.json)"
  rm -f output.json*
}

@test "resource=dashboard_input | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_input" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 9th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "9d" output.json
  else
    run sed -i "9d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_input.json)"
  rm -f output.json*
}

@test "resource=dashboard_table | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  steampipe query "select * from steampipe_dashboard_table" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 9th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "9d" output.json
  else
    run sed -i "9d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_table.json)"
  rm -f output.json*
}

@test "ensure mod name in introspection table is <mod_name> not mod.<mod_name>" {
  cd $SIMPLE_MOD_DIR
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

@test "ensure the reference_from column is populated correctly" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_reference" --output json

  # extract the refs and the referenced_by of the variable `sample_var_1`
  refs=$(echo $output | jq '.[] | select(.reference_to=="var.sample_var_1") | .reference_from')
  echo $refs

  assert_equal "$refs" '"query.sample_query_1"'
}

@test "introspection tables should get populated in query batch mode" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_query" --output json

  # extracting only description from the list, which is enough to prove that there is an output
  description=$(echo $output | jq '.[].description')
  assert_equal "$description" '"query 1 - 3 params all with defaults"'
}

@test "steampipe_introspection=control" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=control
  steampipe query "select * from steampipe_control" --output json > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 11th line, since it contains file location which would differ in github runners
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".json" "11d" output.json
  else
    run sed -i "11d" output.json
  fi

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_info_control.json)"
  rm -f output.json*
}

@test "steampipe check --where | steampipe_introspection=control" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=control
  steampipe check control.sample_control_1 --where "severity in ('high')" --export output.json

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_check_where.json)"
  rm -f output.json*
}

@test "steampipe check --tag | steampipe_introspection=control" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=control
  steampipe check control.sample_control_1 --tag foo=bar --export output.json

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_introspection_check_where.json)"
  rm -f output.json*
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
