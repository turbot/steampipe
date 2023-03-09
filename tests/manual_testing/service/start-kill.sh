for i in {1..10}; do
  echo "############################################################### STARTING"
  STEAMPIPE_LOG=trace steampipe service start
  ps -ef | grep steampipe
  STEAMPIPE_LOG=trace steampipe query "select pg_sleep(10)" &
  
  echo "############################################################### KILLING"
  pkill -9 steampipe
  ps -ef | grep steampipe
  pkill -9 postgres
  ps -ef | grep steampipe
  echo "############################################################### DONE"
done
