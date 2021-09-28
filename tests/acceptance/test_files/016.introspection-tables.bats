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