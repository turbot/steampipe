for i in {1..10}; do
  echo "############################################################### STARTING"
  STEAMPIPE_LOG=trace steampipe service start
  echo "############################################################### STOPPING"
  STEAMPIPE_LOG=trace steampipe service stop
  echo "############################################################### DONE"
done