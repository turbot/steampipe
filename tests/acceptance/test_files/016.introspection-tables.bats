load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

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

@test "ensure the reference_to and reference_from columns are populated correctly" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_reference" --output json

  # extract the refs and the referenced_by
  refs=$(echo $output | jq '.[0].reference_to')
  referenced_by=$(echo $output | jq '.[0].reference_from')

  assert_equal "$refs" '"var.sample_var_1"'
  assert_equal "$referenced_by" '"query.sample_query_1"'
}

@test "ensure the reference_to column includes variable references" {
  cd $SIMPLE_MOD_DIR
  run steampipe query "select * from steampipe_reference" --output json

  # extract the refs
  refs=$(echo $output | jq '.[0].reference_to')
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

  # extracting only description from the list, which is enough to prove that there is an output
  description=$(echo $output | jq '.[].description')
  assert_equal "$description" '"query 1 - 3 params all with defaults"'
}