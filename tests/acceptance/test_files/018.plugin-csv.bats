load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "check csv plugin" {
  run steampipe plugin install csv

  cd $SRC_DATA_DIR
  # appending the csv_plugin_test path
  full_path="${CSV_PATH}/test_data/csv_plugin_test/*.csv"
  echo "${full_path}"

  # escaping the slashes(/)
  b=$(echo -e "${full_path}" | sed -e 's/\//\\\//g')
  echo -e $b

  # reading each line from the config template and storing in a file
  while IFS= read -r line
  do
    echo "$line" >> output.spc
  done < "csv_template.spc"

  # replace the config file template with required path
  sed -i -e "s/abc/${b}/g" 'output.spc'

  # copy the new connection config
  cp output.spc $STEAMPIPE_INSTALL_DIR/config/csv1.spc

  # run the query and verify
  run steampipe query "select * from csv1.csv_file_1"
  assert_success
  rm -f $STEAMPIPE_INSTALL_DIR/config/csv1.spc
  rm -f output.*
}
