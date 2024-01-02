# Function to check if all 'state' values 
# in the steampipe_connection_state stable are "ready"
wait_connection_map_stable() {
    local timeout_duration=5
    local end_time=$(( $(date +%s) + timeout_duration ))
    local all_ready=false

    while [[ $(date +%s) -lt $end_time ]]
    do
      # Run the steampipe query and parse the JSON output
      local json_output=$(steampipe query "select * from steampipe_connection_state" --output json)
      if [ $? -ne 0 ]; then
        echo "Failed to execute steampipe query"
        return 1
      fi

      for state in $(echo $json_output | jq -r '.[].state')
      do
        if [ "$state" != "ready" ]; then
          # wait for sometime 
          sleep 0.5
          # and try again
          continue
        fi
      done
      
      # if we are here that means all are in the ready state 
      all_ready=true
      # we can break out of the loop
      break
    done

    if [ "$all_ready" = true ]; then
      return 0
    else
      return 1
    fi
}


